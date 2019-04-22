package sysinfo

import (
	"os/exec"
	"strconv"
	"strings"
)

var osxVersionMap = map[string]string{
	// OS X versions named after big cats intentionally omitted
	"10.9":  "OS X Mavericks",
	"10.10": "OS X Yosemite",
	"10.11": "OS X El Capitan",
	"10.12": "macOS Sierra",
	"10.13": "macOS High Sierra",
	"10.14": "macOS Mojave",
}

func getPlatformSpecifics(info SystemInfo) SystemInfo {
	info.Name = "UNKNOWN"
	info.Version = "UNKNOWN"
	info.Model = "UNKNOWN"
	info.CPU = "UNKNOWN"
	info.Memory = -1

	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err == nil {
		info.Version = strings.TrimSpace(string(out))
		info.Name = resolveNameFromVersion(info.Version)
	}

	out, err = exec.Command("sysctl", "hw.model").Output()
	if err == nil {
		// The format is something like: "hw.model: MacBookPro12,1"
		info.Model = strings.TrimSpace(strings.Fields(string(out))[1])
	}

	out, err = exec.Command("sysctl", "machdep.cpu.brand_string").Output()
	if err == nil {
		// The format is something like: "machdep.cpu.brand_string: Intel(R) Core(TM) i7-5557U CPU @ 3.10GHz"
		info.CPU = strings.TrimSpace(strings.SplitN(string(out), ":", 2)[1])
	}

	out, err = exec.Command("sysctl", "hw.memsize").Output()
	if err == nil {
		// The format is something like: "hw.memsize: 17179869184"
		memStr := strings.TrimSpace(strings.Fields(string(out))[1])
		info.Memory, _ = strconv.ParseInt(memStr, 10, 64)
	}

	return info
}

func resolveNameFromVersion(version string) string {
	for k, v := range osxVersionMap {
		if strings.HasPrefix(version, k) {
			return v
		}
	}
	return "UNKNOWN"
}
