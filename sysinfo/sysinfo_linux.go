package sysinfo

import (
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

func getPlatformSpecifics(info SystemInfo) SystemInfo {
	info.Name, info.Version = linuxVersion()

	info.Model = "UNKNOWN" // No easy way to find out system details on Linux
	info.CPU = "UNKNOWN"
	info.Memory = -1

	out, err := exec.Command("cat", "/proc/cpuinfo").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			fs := strings.Split(line, ":")
			if len(fs) < 2 {
				continue
			}
			k := strings.TrimSpace(fs[0])
			v := strings.TrimSpace(fs[1])

			if k == "model name" {
				info.CPU = v
				break
			}
		}
	}

	out, err = exec.Command("cat", "/proc/meminfo").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			fs := strings.Split(line, ":")
			if len(fs) < 2 {
				continue
			}
			k := strings.TrimSpace(fs[0])
			v := strings.TrimSpace(fs[1])

			if k == "MemTotal" {
				v = strings.Replace(v, " kB", "", -1)
				n, _ := strconv.ParseInt(v, 10, 64)
				info.Memory = n * 1024
				break
			}
		}
	}

	return info
}

// getLinuxVersion is a bit tedious as we try to figure out the vendor
// and the vendor version and there seem to be a lot of different
// variants for this.
func linuxVersion() (name string, version string) {
	name = "UNKNOWN"
	version = "UNKNOWN"

	content, err := ioutil.ReadFile("/etc/alpine-release")
	if err == nil {
		name = "Alpine"
		version = strings.TrimSpace(string(content[:]))
		return name, version
	}

	// lsb_release is the fallback
	out, err := exec.Command("lsb_release", "-a").Output()
	if err != nil {
		return name, version
	}
	for _, line := range strings.Split(string(out), "\n") {
		fs := strings.Split(line, ":")
		if len(fs) < 2 {
			continue
		}
		k := strings.TrimSpace(fs[0])
		v := strings.TrimSpace(fs[1])
		switch k {
		case "Distributor ID":
			name = v
		case "Release":
			version = v
		}
	}
	return name, version
}
