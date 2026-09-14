package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// captureStderr temporarily redirects the package-level os.Stderr that run()
// writes to directly, so an in-process call can assert on its output.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()

	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

var binPath string

// TestMain builds bg-deadline once and shares the binary across subtests,
// since each subtest exercises real process/signal behavior via subprocess.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "bg-deadline-test-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	binPath = filepath.Join(dir, "bg-deadline")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go build bg-deadline: %v\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func exitCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	t.Fatalf("unexpected error type: %v", err)
	return -1
}

func TestBgDeadline(t *testing.T) {
	t.Run("child_exits_before_deadline", func(t *testing.T) {
		for _, want := range []int{0, 3} {
			t.Run(strconv.Itoa(want), func(t *testing.T) {
				cmd := exec.Command(binPath, "5", "sh", "-c", fmt.Sprintf("exit %d", want))
				start := time.Now()
				err := cmd.Run()
				elapsed := time.Since(start)

				if got := exitCodeOf(t, err); got != want {
					t.Fatalf("exit code = %d, want %d", got, want)
				}
				if elapsed > 3*time.Second {
					t.Fatalf("child exit took too long: %v (deadline should not have fired)", elapsed)
				}
			})
		}
	})

	t.Run("deadline_exceeded_returns_124", func(t *testing.T) {
		cmd := exec.Command(binPath, "1", "sh", "-c", "sleep 30")
		cmd.Env = append(os.Environ(), graceEnvVar+"=1")

		err := cmd.Run()
		if got := exitCodeOf(t, err); got != exitTimeout {
			t.Fatalf("exit code = %d, want %d", got, exitTimeout)
		}
	})

	t.Run("deadline_kills_grandchildren", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "grandchild.pid")
		script := fmt.Sprintf(`sleep 30 & echo $! > %q; sleep 30`, pidFile)

		cmd := exec.Command(binPath, "1", "sh", "-c", script)
		cmd.Env = append(os.Environ(), graceEnvVar+"=1")

		if err := cmd.Run(); err != nil {
			if got := exitCodeOf(t, err); got != exitTimeout {
				t.Fatalf("exit code = %d, want %d", got, exitTimeout)
			}
		}

		pidBytes, err := os.ReadFile(pidFile)
		if err != nil {
			t.Fatalf("grandchild pid file not written: %v", err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
		if err != nil {
			t.Fatalf("grandchild pid file content invalid: %q", pidBytes)
		}

		deadline := time.Now().Add(2 * time.Second)
		for {
			if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("grandchild pid %d still alive after deadline kill", pid)
			}
			time.Sleep(50 * time.Millisecond)
		}
	})

	t.Run("sigterm_ignored_child_is_killed", func(t *testing.T) {
		cmd := exec.Command(binPath, "1", "sh", "-c", `trap '' TERM; sleep 30`)
		cmd.Env = append(os.Environ(), graceEnvVar+"=1")

		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)

		if got := exitCodeOf(t, err); got != exitTimeout {
			t.Fatalf("exit code = %d, want %d", got, exitTimeout)
		}
		// deadline(1s) + grace(1s) should suffice; SIGKILL cannot be trapped.
		if elapsed < 1*time.Second || elapsed > 5*time.Second {
			t.Fatalf("unexpected elapsed time: %v", elapsed)
		}
	})

	t.Run("deadline_message_points_to_monitor", func(t *testing.T) {
		cmd := exec.Command(binPath, "1", "sh", "-c", "sleep 30")
		cmd.Env = append(os.Environ(), graceEnvVar+"=1")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		out := stderr.String()
		if !strings.Contains(out, "Monitor") || !strings.Contains(out, "persistent") {
			t.Fatalf("stderr marker missing Monitor/persistent: %q", out)
		}
	})

	t.Run("orphaned_group_member_is_reaped", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "orphan.pid")
		script := fmt.Sprintf(`sleep 30 & echo $! > %q; exit 0`, pidFile)

		cmd := exec.Command(binPath, "5", "sh", "-c", script)
		cmd.Env = append(os.Environ(), graceEnvVar+"=1")

		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)

		if got := exitCodeOf(t, err); got != 0 {
			t.Fatalf("exit code = %d, want 0 (direct child's own status)", got)
		}
		if elapsed > 3*time.Second {
			t.Fatalf("wrapper took too long to return: %v (deadline should not have fired)", elapsed)
		}

		pidBytes, err := os.ReadFile(pidFile)
		if err != nil {
			t.Fatalf("orphan pid file not written: %v", err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
		if err != nil {
			t.Fatalf("orphan pid file content invalid: %q", pidBytes)
		}

		deadline := time.Now().Add(2 * time.Second)
		for {
			if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("orphaned group member pid %d still alive after wrapper returned", pid)
			}
			time.Sleep(50 * time.Millisecond)
		}
	})

	t.Run("deadline_kills_term_ignoring_grandchild", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "grandchild.pid")
		// SIG_IGN survives exec, so the inner shell's own TERM-ignore also
		// covers the sleep it execs; wrapping it in another sh -c makes it a
		// grandchild of the direct child, not the child itself.
		inner := fmt.Sprintf(`trap "" TERM; echo $$ > %q; sleep 30`, pidFile)
		script := fmt.Sprintf(`sh -c '%s' & wait`, inner)

		cmd := exec.Command(binPath, "1", "sh", "-c", script)
		cmd.Env = append(os.Environ(), graceEnvVar+"=1")

		if err := cmd.Run(); err != nil {
			if got := exitCodeOf(t, err); got != exitTimeout {
				t.Fatalf("exit code = %d, want %d", got, exitTimeout)
			}
		}

		pidBytes, err := os.ReadFile(pidFile)
		if err != nil {
			t.Fatalf("grandchild pid file not written: %v", err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
		if err != nil {
			t.Fatalf("grandchild pid file content invalid: %q", pidBytes)
		}

		deadline := time.Now().Add(3 * time.Second)
		for {
			if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("TERM-ignoring grandchild pid %d still alive after grace SIGKILL", pid)
			}
			time.Sleep(50 * time.Millisecond)
		}
	})

	t.Run("external_sigterm_forwards_then_exits", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "child.pid")
		script := fmt.Sprintf(`echo $$ > %q; sleep 30`, pidFile)

		cmd := exec.Command(binPath, "30", "sh", "-c", script)
		cmd.Env = append(os.Environ(), graceEnvVar+"=1")
		if err := cmd.Start(); err != nil {
			t.Fatalf("start bg-deadline: %v", err)
		}

		var pidBytes []byte
		readDeadline := time.Now().Add(2 * time.Second)
		for {
			b, readErr := os.ReadFile(pidFile)
			if readErr == nil && len(strings.TrimSpace(string(b))) > 0 {
				pidBytes = b
				break
			}
			if time.Now().After(readDeadline) {
				t.Fatalf("child pid file was not written in time")
			}
			time.Sleep(20 * time.Millisecond)
		}
		childPid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
		if err != nil {
			t.Fatalf("child pid file content invalid: %q", pidBytes)
		}

		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatalf("signal bg-deadline: %v", err)
		}

		waitErr := cmd.Wait()
		want := 128 + int(syscall.SIGTERM)
		if got := exitCodeOf(t, waitErr); got != want {
			t.Fatalf("exit code = %d, want %d", got, want)
		}
		if !errors.Is(syscall.Kill(childPid, 0), syscall.ESRCH) {
			t.Fatalf("child pid %d still alive after wrapper forwarded SIGTERM", childPid)
		}
	})

	t.Run("child_exit_at_deadline_is_not_misreported", func(t *testing.T) {
		for i := range 5 {
			cmd := exec.Command(binPath, "1", "sh", "-c", "sleep 0.9; exit 7")
			cmd.Env = append(os.Environ(), graceEnvVar+"=1")

			err := cmd.Run()
			if got := exitCodeOf(t, err); got != 7 {
				t.Fatalf("iteration %d: exit code = %d, want 7 (child exited on its own near the deadline, not a real timeout)", i, got)
			}
		}
	})

	t.Run("grace_env_is_clamped", func(t *testing.T) {
		cases := []struct {
			in   string
			want time.Duration
		}{
			{"", defaultGrace},
			{"0", defaultGrace},
			{"-5", defaultGrace},
			{"10000", defaultGrace},
			{"abc", defaultGrace},
			{strconv.Itoa(minGraceSeconds), time.Duration(minGraceSeconds) * time.Second},
			{strconv.Itoa(maxGraceSeconds), time.Duration(maxGraceSeconds) * time.Second},
			{"3", 3 * time.Second},
		}
		for _, c := range cases {
			if got := graceFromEnv(c.in); got != c.want {
				t.Errorf("graceFromEnv(%q) = %v, want %v", c.in, got, c.want)
			}
		}
	})

	t.Run("child_is_not_reaped_before_group_cleanup", func(t *testing.T) {
		// Calls run() in-process (rather than via the built binary) so the
		// hook below can observe the child's pid state from inside the
		// cleanup window, which a subprocess can't see from the outside.
		prevHook := beforeGroupCleanup
		defer func() { beforeGroupCleanup = prevHook }()

		var (
			hookCalled  bool
			observedPid int
		)
		beforeGroupCleanup = func(pgid int) {
			hookCalled = true
			observedPid = pgid
			if err := syscall.Kill(pgid, 0); err != nil {
				t.Errorf("child pid %d not held as a zombie during group cleanup: %v", pgid, err)
			}
		}

		got := run([]string{"5", "sh", "-c", "exit 0"})

		if !hookCalled {
			t.Fatal("beforeGroupCleanup hook was never invoked")
		}
		if got != 0 {
			t.Fatalf("exit code = %d, want 0", got)
		}
		if !errors.Is(syscall.Kill(observedPid, 0), syscall.ESRCH) {
			t.Fatalf("child pid %d still present after run() returned (not reaped)", observedPid)
		}
	})

	t.Run("observer_failure_refuses_to_run", func(t *testing.T) {
		// Forces the exit-observer constructor to fail; the wrapper must
		// refuse to run the command further rather than race an unreliable
		// wait, reporting exitObserverFailure instead.
		prevNewObserver := newExitObserverFn
		defer func() { newExitObserverFn = prevNewObserver }()
		newExitObserverFn = func(int) (*exitObserver, error) { return nil, fmt.Errorf("forced failure for test") }
		t.Setenv(graceEnvVar, "1")

		var got int
		stderr := captureStderr(t, func() {
			got = run([]string{"5", "sh", "-c", "exit 0"})
		})

		if got != exitObserverFailure {
			t.Fatalf("exit code = %d, want %d", got, exitObserverFailure)
		}
		if !strings.Contains(stderr, "refusing to run unbounded") {
			t.Fatalf("stderr missing 'refusing to run unbounded': %q", stderr)
		}
	})

	t.Run("observer_failure_kills_grandchildren", func(t *testing.T) {
		// Same forced-failure seam as above; even though the observer never
		// got to run, the fail-closed path must still clean up anything the
		// direct child left behind in its process group before reaping it.
		pidFile := filepath.Join(t.TempDir(), "grandchild.pid")

		prevNewObserver := newExitObserverFn
		defer func() { newExitObserverFn = prevNewObserver }()
		newExitObserverFn = func(int) (*exitObserver, error) {
			// The fail-closed kill is near-instant; wait for the grandchild
			// to actually exist first so the test exercises the intended
			// "leftover backgrounded process" scenario instead of racing it.
			deadline := time.Now().Add(2 * time.Second)
			for {
				if b, err := os.ReadFile(pidFile); err == nil && len(strings.TrimSpace(string(b))) > 0 {
					break
				}
				if time.Now().After(deadline) {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			return nil, fmt.Errorf("forced failure for test")
		}
		t.Setenv(graceEnvVar, "1")

		script := fmt.Sprintf(`sleep 30 & echo $! > %q; exit 0`, pidFile)

		got := run([]string{"5", "sh", "-c", script})
		if got != exitObserverFailure {
			t.Fatalf("exit code = %d, want %d", got, exitObserverFailure)
		}

		pidBytes, err := os.ReadFile(pidFile)
		if err != nil {
			t.Fatalf("grandchild pid file not written: %v", err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
		if err != nil {
			t.Fatalf("grandchild pid file content invalid: %q", pidBytes)
		}
		if !errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
			t.Fatalf("grandchild pid %d still alive after observer-failure cleanup", pid)
		}
	})

	t.Run("kevent_ev_error_is_treated_as_observer_failure", func(t *testing.T) {
		ev := syscall.Kevent_t{
			Flags: syscall.EV_ERROR,
			Data:  int64(syscall.ESRCH),
		}
		if err := exitEventResult(ev); !errors.Is(err, syscall.ESRCH) {
			t.Fatalf("exitEventResult(EV_ERROR) = %v, want ESRCH", err)
		}
	})

	t.Run("kevent_registration_is_retried_without_losing_changelist", func(t *testing.T) {
		attempts := 0
		var seenChanges [][]syscall.Kevent_t
		call := func(changes, events []syscall.Kevent_t, timeout *syscall.Timespec) (int, error) {
			attempts++
			seenChanges = append(seenChanges, changes)
			if attempts < 3 {
				return 0, syscall.EINTR
			}
			return 1, nil
		}
		changes := []syscall.Kevent_t{{Ident: 123}}
		events := make([]syscall.Kevent_t, 1)

		n, err := retryKeventOnEINTR(call, changes, events, &syscall.Timespec{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 1 {
			t.Fatalf("n = %d, want 1", n)
		}
		if attempts != 3 {
			t.Fatalf("attempts = %d, want 3", attempts)
		}
		for i, c := range seenChanges {
			if len(c) == 0 || c[0].Ident != 123 {
				t.Fatalf("attempt %d: changelist was lost/emptied: %+v", i, c)
			}
		}
	})
}
