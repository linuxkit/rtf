//go:build !linux && !darwin && !windows
// +build !linux,!darwin,!windows

package sysinfo

func getPlatformSpecifics(info SystemInfo) SystemInfo {
	return info
}
