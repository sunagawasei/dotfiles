// Command zmk-build builds the roBa ZMK firmware locally in an Apple
// `container` (zmk-build-arm image) and flashes .uf2 files to a keyboard half
// that is in bootloader mode.
//
//	zmk-build [build] [-repo DIR] [target...]   targets: R L reset (default: all)
//	zmk-build flash [-repo DIR] <R|L|reset>
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type target struct {
	key, shield, snippet string
}

// Same set as the former CI matrix; roBa_R carries the ZMK Studio snippet.
var targets = []target{
	{"R", "roBa_R", "studio-rpc-usb-uart"},
	{"L", "roBa_L", ""},
	{"reset", "settings_reset", ""},
}

const (
	image  = "zmkfirmware/zmk-build-arm:stable"
	board  = "seeeduino_xiao_ble"
	script = `#!/bin/bash
set -euo pipefail
shield=$1; snippet=${2:-}
mkdir -p /ws/config && rm -rf /ws/config/* && cp -R /src/config/* /ws/config/
cd /ws
[ -d .west ] || west init -l /ws/config
west update --fetch-opt=--filter=tree:0
west zephyr-export
args=(-s zmk/app -d "/ws/build/$shield" -b ` + board + `)
[ -n "$snippet" ] && args+=(-S "$snippet")
west build "${args[@]}" -- -DZMK_CONFIG=/ws/config -DSHIELD="$shield" -DZMK_EXTRA_MODULES=/src
cp "/ws/build/$shield/zephyr/zmk.uf2" "/out/$shield-` + board + `-zmk.uf2"
`
)

func find(key string) (target, bool) {
	for _, t := range targets {
		if t.key == key {
			return t, true
		}
	}
	return target{}, false
}

func uf2Name(t target) string { return t.shield + "-" + board + "-zmk.uf2" }

func main() {
	args := os.Args[1:]
	cmd := "build"
	if len(args) > 0 && (args[0] == "build" || args[0] == "flash") {
		cmd, args = args[0], args[1:]
	}
	home, _ := os.UserHomeDir()
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	repo := fs.String("repo", filepath.Join(home, "poc", "zmk-config-roBa"), "zmk-config repo")
	cache := fs.String("cache", filepath.Join(home, ".cache", "zmk-roBa"), "west workspace cache")
	fs.Parse(args)

	var err error
	switch cmd {
	case "build":
		err = build(*repo, *cache, fs.Args())
	case "flash":
		err = flash(*repo, fs.Args())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "zmk-build:", err)
		os.Exit(1)
	}
}

func build(repo, cache string, keys []string) error {
	sel := targets
	if len(keys) > 0 {
		sel = nil
		for _, k := range keys {
			t, ok := find(k)
			if !ok {
				return fmt.Errorf("unknown target %q (R, L, reset)", k)
			}
			sel = append(sel, t)
		}
	}
	out := filepath.Join(cache, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cache, "build-in.sh"), []byte(script), 0o755); err != nil {
		return err
	}
	for _, t := range sel {
		fmt.Println("== build", t.shield)
		c := exec.Command("container", "run", "--rm", "-c", "4", "-m", "6G",
			"-v", cache+":/ws", "-v", repo+":/src:ro", "-v", out+":/out",
			image, "bash", "/ws/build-in.sh", t.shield, t.snippet)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("%s: %w", t.shield, err)
		}
		if err := copyFile(filepath.Join(out, uf2Name(t)), filepath.Join(repo, "firmware", uf2Name(t))); err != nil {
			return err
		}
		fmt.Println("ok", filepath.Join(repo, "firmware", uf2Name(t)))
	}
	return nil
}

func flash(repo string, keys []string) error {
	if len(keys) != 1 {
		return fmt.Errorf("usage: zmk-build flash <R|L|reset>")
	}
	t, ok := find(keys[0])
	if !ok {
		return fmt.Errorf("unknown target %q (R, L, reset)", keys[0])
	}
	src := filepath.Join(repo, "firmware", uf2Name(t))
	if _, err := os.Stat(src); err != nil {
		return err
	}
	fmt.Println("bootloaderモードのキーボード待ち(リセットを2回押す)...")
	vol := waitVolume(true)
	fmt.Println("検出:", vol)
	// UF2 bootloader reboots as soon as the copy finishes, so a copy error
	// (xattr failure) after the data is written is expected.
	if err := copyFile(src, filepath.Join(vol, uf2Name(t))); err != nil {
		fmt.Fprintln(os.Stderr, "copy:", err)
	}
	waitVolume(false)
	fmt.Println(t.shield, "の書き込み完了")
	return nil
}

// waitVolume blocks until a bootloader volume appears (present) or vanishes.
func waitVolume(present bool) string {
	for {
		for _, n := range []string{"XIAO-SENSE", "NO NAME"} {
			p := filepath.Join("/Volumes", n)
			if _, err := os.Stat(p); err == nil && present {
				return p
			}
		}
		if !present {
			gone := true
			for _, n := range []string{"XIAO-SENSE", "NO NAME"} {
				if _, err := os.Stat(filepath.Join("/Volumes", n)); err == nil {
					gone = false
				}
			}
			if gone {
				return ""
			}
		}
		time.Sleep(time.Second)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
