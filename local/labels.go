package local

import (
	"fmt"
	"strconv"
	"strings"
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

// WillRun determines if a group or test should run based on its labels and the RunConfig
func WillRun(labels, notLabels map[string]bool, config RunConfig) bool {
	// 2. Check every test label is in the hostLabels
	for l := range labels {
		if _, ok := config.Labels[l]; !ok {
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

// CheckPattern implements a test to see if a given name matches the provided pattern
func CheckPattern(name, pattern string) bool {
	// 1. Check that name begins with the TestPattern
	if !strings.HasPrefix(name, pattern) {
		return false
	}
	return true
}

func makeLabelString(labels map[string]bool, notLabels map[string]bool) string {
	var l []string
	for s := range notLabels {
		l = append(l, fmt.Sprintf("!%s", s))
	}
	for s := range labels {
		l = append(l, s)
	}
	return strings.Join(l, ", ")
}

func getNameAndOrder(path string) (int, string) {
	parts := strings.SplitN(path, "_", 2)
	if len(parts) < 2 {
		return 0, path
	}
	order, _ := strconv.Atoi(parts[0])
	return order, parts[1]
}
