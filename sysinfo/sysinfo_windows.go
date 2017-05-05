package sysinfo

import (
	"encoding/json"
	"os/exec"
)

func getPlatformSpecifics(info SystemInfo) SystemInfo {
	out, err := exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		"&{",
		"$(Get-WmiObject Win32_OperatingSystem) | ConvertTo-json -Depth 1",
		"}",
	).Output()
	if err != nil {
		info.Name = "UNKNOWN"
		info.Version = "UNKNOWN"
		return info
	}
	var st interface{}
	json.Unmarshal([]byte(out), &st)
	osMap := st.(map[string]interface{})

	info.Name = osMap["Caption"].(string)
	info.Version = osMap["Version"].(string)

	return info
}
