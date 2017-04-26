package sysinfo

import "runtime"

// SystemInfo encapsulates useful information about a system
type SystemInfo struct {
	OS      string
	Name    string
	Version string
	Arch    string
}

// GetSystemInfo populates a new SystemInfo with both generic and
// platform specific information
func GetSystemInfo() SystemInfo {
	info := SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
	return getPlatformSpecifics(info)
}

// List returns the system info as a list of strings
func (s SystemInfo) List() []string {
	l := []string{
		s.OS,
		s.Name,
		s.Version,
		s.Arch,
	}
	if s.OS == "darwin" {
		l = append(l, "osx")
	}
	if s.OS == "windows" {
		l = append(l, "win")
	}
	return l
}
