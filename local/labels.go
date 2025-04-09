package local

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/linuxkit/rtf/sysinfo"
)

// ParseLabels constucts a map[string]bool for both positive and negative labels from a comma separated list
func ParseLabels(labels string) (map[string]bool, map[string]bool) {
	set := make(map[string]bool)
	unSet := make(map[string]bool)

	if labels == "" {
		return set, unSet
	}

	l := strings.Split(labels, ",")
	for _, s := range l {
		if strings.HasPrefix(s, "!") {
			unSet[s[1:]] = true
		} else {
			set[s] = true
		}
	}
	return set, unSet
}

// CheckLabel determines if a group or test should run based on its labels and the RunConfig
func CheckLabel(labels, notLabels map[string]bool, config RunConfig) bool {
	// 1. If test has labels
	if len(labels) > 0 {
		// 2. Check that at least one test label is in the hostLabels
		matches := 0
		for l := range labels {
			if _, ok := config.Labels[l]; ok {
				matches++
			}
		}
		if matches == 0 {
			return false
		}
	}

	// 3. Check every test notLabel is NOT in the hostLabels
	for l := range notLabels {
		if _, ok := config.Labels[l]; ok {
			return false
		}
	}
	// 4. Check that none of the test labels appear in the hostNotLabels
	for l := range config.NotLabels {
		if _, ok := labels[l]; ok {
			return false
		}
	}
	return true
}

func makeLabelString(labels map[string]bool, notLabels map[string]bool, sep string) string {
	var l []string
	for s := range notLabels {
		l = append(l, fmt.Sprintf("!%s", s))
	}
	for s := range labels {
		l = append(l, s)
	}
	return strings.Join(l, sep)
}

func getNameAndOrder(path string) (int, string) {
	parts := strings.SplitN(path, "_", 2)
	if len(parts) < 2 {
		return 0, path
	}
	order, _ := strconv.Atoi(parts[0])
	return order, parts[1]
}

func applySystemLabels(labels string) (map[string]bool, map[string]bool) {
	systemInfo := sysinfo.GetSystemInfo()
	l, nl := ParseLabels(labels)
	for _, v := range systemInfo.List() {
		if _, ok := l[v]; !ok {
			l[v] = true
		}
	}
	return l, nl
}

// NewRunConfig returns a new RunConfig from test labels and a pattern
func NewRunConfig(labels string, pattern string) RunConfig {
	matchedLabels, notLabels := applySystemLabels(labels)
	return RunConfig{
		TestPattern: pattern,
		Labels:      matchedLabels,
		NotLabels:   notLabels,
	}
}

// ValidatePattern validates that an arg string is a valid test pattern
func ValidatePattern(args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("expected only one test pattern")
	}
	if len(args) == 0 {
		return "", nil
	}
	return args[0], nil
}
