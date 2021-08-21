package constants

import (
	"os"
	"syscall"
	"time"

	"github.com/gepis/strge/pkg/mount"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func EnsureRemoveAll(dir string) error {
	notExistErr := make(map[string]bool)

	// track retries
	exitOnErr := make(map[string]int)
	maxRetry := 100

	// Attempt to unmount anything beneath this dir first
	if err := mount.RecursiveUnmount(dir); err != nil {
		logrus.Debugf("RecusiveUnmount on %s failed: %v", dir, err)
	}

	for {
		err := os.RemoveAll(dir)
		if err == nil {
			return nil
		}

		pe, ok := err.(*os.PathError)
		if !ok {
			return err
		}

		if os.IsNotExist(err) {
			if notExistErr[pe.Path] {
				return err
			}

			notExistErr[pe.Path] = true

			// There is a race where some subdir can be removed but after the parent
			//   dir entries have been read.
			// So the path could be from `os.Remove(subdir)`
			// If the reported non-existent path is not the passed in `dir` we
			// should just retry, but otherwise return with no error.
			if pe.Path == dir {
				return nil
			}

			continue
		}

		if pe.Err != syscall.EBUSY {
			return err
		}

		if e := mount.Unmount(pe.Path); e != nil {
			return errors.Wrapf(e, "error while removing %s", dir)
		}

		if exitOnErr[pe.Path] == maxRetry {
			return err
		}

		exitOnErr[pe.Path]++
		time.Sleep(100 * time.Millisecond)
	}
}
