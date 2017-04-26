package sysinfo

import (
	"os/exec"
	"strings"
)

var osxVersionMap = map[string]string{
	// OS X versions named after big cats intentionally omitted
	"10.9":  "OS X Mavericks",
	"10.10": "OS X Yosemite",
	"10.11": "OS X El Capitan",
	"10.12": "macOS Sierra",
}

func getPlatformSpecifics(info SystemInfo) SystemInfo {
	out, err := exec.Command("sw_vers-pver", "sw_vers", "-productVersion").Output()
	if err != nil {
		info.Name = "UNKNOWN"
		info.Version = "UNKNOWN"
		return info
	}
	info.Version = strings.TrimSpace(string(out))
	var ok bool
	info.Name, ok = osxVersionMap[info.Version]
	if !ok {
		info.Version = "UNKNOWN"
	}
	return info
}
