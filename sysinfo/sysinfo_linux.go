package sysinfo

import (
	"io/ioutil"
	"strings"
)

func getPlatformSpecifics(info SystemInfo) SystemInfo {
	info.Name, info.Version = linuxVersion()
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
	c := Cmd{SystemDir + "lsb_release", "lsb_release", []string{"-a"}, 0}
	out := c.Gather(ctx, w)
	for _, line := range strings.Split(out, "\n") {
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
