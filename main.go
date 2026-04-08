package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Boyarinov/pomp/internal/recent"
	"github.com/Boyarinov/pomp/internal/ui"
)

const usage = `pomp — pre-omp project picker

Usage:
  pomp            Launch the picker and run omp in the selected directory.
  pomp -h|--help  Show this help.

Environment:
  POMP_SESSIONS_DIR  Override path to omp sessions dir (default: ~/.omp/agent/sessions).
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			fmt.Print(usage)
			return
		default:
			fmt.Fprintln(os.Stderr, "pomp: unknown argument:", os.Args[1])
			fmt.Fprint(os.Stderr, usage)
			os.Exit(2)
		}
	}

	sessionsDir := resolveSessionsDir()
	entries, err := recent.Scan(sessionsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pomp: failed to scan sessions:", err)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pomp: failed to get cwd:", err)
		os.Exit(1)
	}

	res, err := ui.Run(entries, cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pomp:", err)
		os.Exit(1)
	}
	if res.Selected == "" {
		return
	}

	os.Exit(execOmp(res.Selected))
}

func resolveSessionsDir() string {
	if v := os.Getenv("POMP_SESSIONS_DIR"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".omp", "agent", "sessions")
}

func execOmp(dir string) int {
	cmd := exec.Command("omp")
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "pomp: failed to run omp:", err)
		return 1
	}
	return 0
}
