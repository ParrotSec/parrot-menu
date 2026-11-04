package runlock

import (
	"fmt"
	"os"
	"syscall"
)

const lockPath = "/run/lock/parrot-menu-launcher-updater.lock"

type Lock struct {
	file *os.File
}

func Acquire() (*Lock, error) {
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open updater lock %s: %w", lockPath, err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("lock updater file %s: %w", lockPath, err)
	}
	return &Lock{file: file}, nil
}

func (lock *Lock) Release() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	if err := syscall.Flock(int(lock.file.Fd()), syscall.LOCK_UN); err != nil {
		_ = lock.file.Close()
		return fmt.Errorf("unlock updater file %s: %w", lockPath, err)
	}
	if err := lock.file.Close(); err != nil {
		return fmt.Errorf("close updater lock %s: %w", lockPath, err)
	}
	lock.file = nil
	return nil
}
