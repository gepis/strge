package storage

import (
	"github.com/gepis/strge/pkg/lockfile"
)

type Locker = lockfile.Locker

func GetLockfile(path string) (lockfile.Locker, error) {
	return lockfile.GetLockfile(path)
}

func GetROLockfile(path string) (lockfile.Locker, error) {
	return lockfile.GetROLockfile(path)
}
