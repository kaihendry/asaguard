//go:build windows

package updater

import (
	"os"
	"os/exec"
)

func reexec(exe string) error {
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
