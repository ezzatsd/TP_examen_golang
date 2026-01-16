//go:build darwin

package main

import (
	"syscall"
	"time"
)

func getCreationTime(path string) (time.Time, bool) {
	var stat syscall.Stat_t
	if err := syscall.Stat(path, &stat); err != nil {
		return time.Time{}, false
	}
	sec := stat.Birthtimespec.Sec
	nsec := stat.Birthtimespec.Nsec
	return time.Unix(sec, nsec), true
}
