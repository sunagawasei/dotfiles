// Command bg-deadline runs a command under a hard wall-clock deadline and
// reaps its whole process group — not just the direct child — whenever the
// direct child finishes, the deadline fires, or the wrapper itself is asked
// to terminate. This mirrors coreutils `timeout -k` on systems where it
// isn't installed, and additionally guards against a child that backgrounds
// work and exits early, which would otherwise leak an unbounded process.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	exitUsage           = 2
	exitStartFailure    = 1
	exitTimeout         = 124
	exitObserverFailure = 125

	defaultGrace    = 5 * time.Second
	minGraceSeconds = 1
	maxGraceSeconds = 60
	graceEnvVar     = "BG_DEADLINE_GRACE_SECONDS"

	groupPollInterval = 50 * time.Millisecond
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: bg-deadline <deadline_seconds> <command> [args...]")
		return exitUsage
	}

	deadlineSec, err := strconv.Atoi(args[0])
	if err != nil || deadlineSec <= 0 {
		fmt.Fprintf(os.Stderr, "bg-deadline: invalid deadline_seconds %q\n", args[0])
		return exitUsage
	}
	grace := graceFromEnv(os.Getenv(graceEnvVar))

	// Subscribe before Start(): a signal arriving in the gap between Start()
	// and Notify() would otherwise be lost, killing the wrapper without ever
	// forwarding it to the child's group.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer signal.Stop(sigCh)

	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// New process group so a deadline kill reaches grandchildren too; if the
	// kernel can't set it up, Start() fails and we abort instead of letting
	// the child run unbounded outside our control.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "bg-deadline: failed to start command in its own process group: %v\n", err)
		return exitStartFailure
	}
	pgid := cmd.Process.Pid // Setpgid with Pgid=0 makes the child its own group leader.
	reap := func() error { return cmd.Wait() }

	// Observing the child's exit without reaping it is what makes the group
	// cleanup below safe to target at -pgid: as long as the child stays a
	// zombie, its pid/pgid can't be recycled by the kernel for an unrelated
	// process. If we can't set that observer up at all, we can no longer
	// promise the deadline/cleanup contract for this run, so we refuse to
	// run further rather than race an unreliable wait.
	observer, obsErr := newExitObserverFn(pgid)
	if obsErr != nil {
		return failClosed(pgid, grace, reap, obsErr)
	}

	waitCh := make(chan error, 1)
	go func() { waitCh <- observer.wait() }()

	timer := time.NewTimer(time.Duration(deadlineSec) * time.Second)
	defer timer.Stop()

	for {
		select {
		case waitErr := <-waitCh:
			if waitErr != nil {
				return failClosed(pgid, grace, reap, waitErr)
			}
			terminateGroup(pgid, syscall.SIGTERM, grace)
			return exitCodeFromWait(reap())

		case sig := <-sigCh:
			s := sig.(syscall.Signal)
			// Forwarded signal is itself the first kill; only the escalation
			// after grace uses SIGKILL. The wrapper exits here rather than
			// waiting on the deadline, matching how a plain foreground
			// process would react to being signaled.
			terminateGroup(pgid, s, grace)
			_ = reap()
			return 128 + int(s)

		case <-timer.C:
			// select() can ready both timer.C and waitCh at once; check
			// waitCh non-blockingly first so a child that finished right at
			// the deadline is reported as a normal exit, not a spurious
			// timeout.
			select {
			case waitErr := <-waitCh:
				if waitErr != nil {
					return failClosed(pgid, grace, reap, waitErr)
				}
				terminateGroup(pgid, syscall.SIGTERM, grace)
				return exitCodeFromWait(reap())
			default:
			}

			fmt.Fprintf(os.Stderr,
				"bg-deadline: deadline exceeded after %ds; killed. 10分を超える待機は Monitor ツール(timeout_ms 最大3600000 または persistent:true)で流すこと。\n",
				deadlineSec)
			terminateGroup(pgid, syscall.SIGTERM, grace)
			_ = reap()
			return exitTimeout
		}
	}
}

// failClosed handles an exit-observer failure (setup-time or mid-wait): we
// can no longer promise the deadline/cleanup contract for this run, so
// instead of racing a wait we can't reliably interrupt, kill the group and
// refuse to run further. The direct child has not been reaped at this
// point, so its pid/pgid is still reserved and safe to signal.
func failClosed(pgid int, grace time.Duration, reap func() error, reason error) int {
	fmt.Fprintf(os.Stderr,
		"bg-deadline: exit observer unavailable (%v); refusing to run unbounded. killing the process group.\n",
		reason)
	terminateGroup(pgid, syscall.SIGTERM, grace)
	_ = reap()
	return exitObserverFailure
}

// newExitObserverFn constructs the exit observer; overridable in tests to
// exercise the fail-closed path without needing a genuinely broken kqueue.
var newExitObserverFn = newExitObserver

// exitObserver watches a pid's exit via kqueue's EVFILT_PROC/NOTE_EXIT
// without reaping it (the observe-without-reap primitive Darwin actually
// honors: wait4's WNOWAIT flag, despite being accepted, silently reaps the
// child anyway on this OS).
type exitObserver struct {
	kq         int
	events     []syscall.Kevent_t
	firedOnAdd bool
	firedEvent syscall.Kevent_t
}

// newExitObserver creates the kqueue and registers the watch. Registration
// uses a zero timeout so a setup failure (or an already-exited pid) is
// discovered synchronously here, before any select loop starts relying on
// wait() to eventually unblock it.
func newExitObserver(pid int) (*exitObserver, error) {
	kq, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	changes := []syscall.Kevent_t{{
		Ident:  uint64(pid),
		Filter: syscall.EVFILT_PROC,
		Flags:  syscall.EV_ADD | syscall.EV_ONESHOT,
		Fflags: syscall.NOTE_EXIT,
	}}
	events := make([]syscall.Kevent_t, 1)

	call := func(changes, events []syscall.Kevent_t, timeout *syscall.Timespec) (int, error) {
		return syscall.Kevent(kq, changes, events, timeout)
	}
	n, err := retryKeventOnEINTR(call, changes, events, &syscall.Timespec{})
	if err != nil {
		_ = syscall.Close(kq)
		return nil, err
	}

	ob := &exitObserver{kq: kq, events: events}
	if n > 0 {
		ob.firedOnAdd = true
		ob.firedEvent = events[0]
	}
	return ob, nil
}

// wait blocks until the exit event arrives, or returns immediately if it had
// already fired during registration (the pid had already exited by then).
func (o *exitObserver) wait() error {
	defer syscall.Close(o.kq)

	if o.firedOnAdd {
		return exitEventResult(o.firedEvent)
	}
	for {
		n, err := syscall.Kevent(o.kq, nil, o.events, nil)
		if err == syscall.EINTR {
			continue
		}
		if err != nil {
			return err
		}
		if n > 0 {
			return exitEventResult(o.events[0])
		}
	}
}

// retryKeventOnEINTR resends the same changelist on every attempt (not just
// the first) so an interrupt before it is committed can't silently drop the
// registration and leave the caller waiting on an empty kqueue.
func retryKeventOnEINTR(call func(changes, events []syscall.Kevent_t, timeout *syscall.Timespec) (int, error), changes, events []syscall.Kevent_t, timeout *syscall.Timespec) (int, error) {
	for {
		n, err := call(changes, events, timeout)
		if err == syscall.EINTR {
			continue
		}
		return n, err
	}
}

// exitEventResult interprets a kevent returned for our EVFILT_PROC/NOTE_EXIT
// registration. syscall.Kevent is a thin syscall wrapper: it does not turn a
// changelist entry that failed (EV_ERROR) into a Go error, so a plain nil
// error with n>0 does not by itself mean the process actually exited.
func exitEventResult(ev syscall.Kevent_t) error {
	if ev.Flags&syscall.EV_ERROR != 0 {
		return syscall.Errno(ev.Data)
	}
	if ev.Fflags&syscall.NOTE_EXIT == 0 {
		return fmt.Errorf("unexpected kevent (flags=%#x fflags=%#x)", ev.Flags, ev.Fflags)
	}
	return nil
}

// beforeGroupCleanup is a test seam: production code never overrides it.
var beforeGroupCleanup = func(pgid int) {}

// terminateGroup sends first to the whole process group, then escalates to
// SIGKILL only if a member is still alive once grace elapses. Every signal
// send is guarded by a fresh aliveness check: an already-empty pgid may have
// been recycled by the kernel, so we must not blindly signal a number we no
// longer own. Aliveness is polled rather than checked once so a group that
// dies well inside grace doesn't force the wrapper to wait out the full
// duration regardless.
func terminateGroup(pgid int, first syscall.Signal, grace time.Duration) {
	beforeGroupCleanup(pgid) // test hook: observe pgid before any reap could have happened
	if !groupAlive(pgid) {
		return
	}
	_ = syscall.Kill(-pgid, first)

	deadline := time.Now().Add(grace)
	for groupAlive(pgid) && time.Now().Before(deadline) {
		time.Sleep(groupPollInterval)
	}

	if groupAlive(pgid) {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
}

// groupAlive reports whether pgid still has any member. On Darwin,
// kill(-pgid, 0) returns EPERM rather than ESRCH once the only remaining
// member is an unreaped zombie leader, so both errors mean "nothing left to
// signal"; only a nil error means a live process answered.
func groupAlive(pgid int) bool {
	return syscall.Kill(-pgid, 0) == nil
}

// graceFromEnv keeps the kill escalation delay bounded: an unset, unparsable,
// or out-of-window value falls back to defaultGrace rather than letting a
// stray environment value turn "hard deadline" into an unbounded wait.
func graceFromEnv(v string) time.Duration {
	secs, err := strconv.Atoi(v)
	if err != nil || secs < minGraceSeconds || secs > maxGraceSeconds {
		return defaultGrace
	}
	return time.Duration(secs) * time.Second
}

func exitCodeFromWait(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if ws.Signaled() {
				return 128 + int(ws.Signal())
			}
			return ws.ExitStatus()
		}
		return exitErr.ExitCode()
	}
	return exitStartFailure
}
