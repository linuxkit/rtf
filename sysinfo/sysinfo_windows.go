package sysinfo

import (
	"encoding/json"
	"os/exec"
)

func getPlatformSpecifics(info SystemInfo) SystemInfo {
	info.Name = "UNKNOWN"
	info.Version = "UNKNOWN"
	info.Model = "UNKNOWN"
	info.CPU = "UNKNOWN"
	info.Memory = -1

	out, err := exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		"&{",
		"$(Get-WmiObject Win32_OperatingSystem) | ConvertTo-json -Depth 1",
		"}",
	).Output()
	if err == nil {
		var st interface{}
		json.Unmarshal([]byte(out), &st)
		osMap := st.(map[string]interface{})

		info.Name = osMap["Caption"].(string)
		info.Version = osMap["Version"].(string)
	}

	out, err = exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		"&{",
		"$(Get-WmiObject Win32_Processor) | ConvertTo-json -Depth 1",
		"}",
	).Output()
	if err == nil {
		var st interface{}
		json.Unmarshal([]byte(out), &st)
		procMap := st.(map[string]interface{})
		info.CPU = procMap["Name"].(string)
	}

	out, err = exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		"&{",
		"$(Get-WmiObject Win32_ComputerSystem) | ConvertTo-json -Depth 1",
		"}",
	).Output()
	if err == nil {
		var st interface{}
		json.Unmarshal([]byte(out), &st)
		csMap := st.(map[string]interface{})

		info.Model = csMap["Manufacturer"].(string) + " " + csMap["Model"].(string)
		info.Memory = int64(csMap["TotalPhysicalMemory"].(float64))
	}

	return info
}
