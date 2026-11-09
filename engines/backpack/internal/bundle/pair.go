// Package bundle stages and installs the two matching sober-tunnel executables.
package bundle

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func InstallPair(mainSource, helperSource, mainTarget, helperTarget string) error {
	return installPair(mainSource, helperSource, mainTarget, helperTarget, os.Rename)
}
func installPair(mainSource, helperSource, mainTarget, helperTarget string, rename func(string, string) error) error {
	if filepath.Dir(mainTarget) != filepath.Dir(helperTarget) {
		return errors.New("bundle targets must share a directory")
	}
	dir, err := os.MkdirTemp(filepath.Dir(mainTarget), ".sober-bundle-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	for _, file := range []struct{ source, name string }{{mainSource, "next-main"}, {helperSource, "next-helper"}} {
		if err = copyExecutable(file.source, filepath.Join(dir, file.name)); err != nil {
			return fmt.Errorf("staging bundle: %w", err)
		}
	}
	hadHelper := false
	if _, err = os.Stat(helperTarget); err == nil {
		hadHelper = true
		if err = copyExecutable(helperTarget, filepath.Join(dir, "old-helper")); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = rename(filepath.Join(dir, "next-helper"), helperTarget); err != nil {
		return err
	}
	if err = rename(filepath.Join(dir, "next-main"), mainTarget); err != nil {
		var rollback error
		if hadHelper {
			rollback = rename(filepath.Join(dir, "old-helper"), helperTarget)
		} else {
			rollback = os.Remove(helperTarget)
		}
		if rollback != nil {
			return fmt.Errorf("main installation failed (%v); helper rollback failed: %w", err, rollback)
		}
		return fmt.Errorf("main installation failed; helper restored: %w", err)
	}
	return nil
}
func copyExecutable(source, dest string) error {
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > 256<<20 {
		return errors.New("invalid executable file size or type")
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, f); err != nil {
		out.Close()
		return err
	}
	if err = out.Sync(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
