package sysinfo

import "testing"

func TestResolveNameFromVersion(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out string
	}{
		{"10.9", "OS X Mavericks"},
		{"10.10.3", "OS X Yosemite"},
		{"10.11.1", "OS X El Capitan"},
		{"10.12.6", "macOS Sierra"},
		{"10.13", "macOS High Sierra"},
	} {
		if name := resolveNameFromVersion(tt.in); name != tt.out {
			t.Errorf("resolveNameFromVersion %s\nExpected: %s\nGot: %s\n", tt.in, tt.out, name)
		}
	}
}
