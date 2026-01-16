//go:build !darwin

package main

import "time"

func getCreationTime(path string) (time.Time, bool) {
	return time.Time{}, false
}
