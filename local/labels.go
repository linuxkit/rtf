package local

import (
	"fmt"
	"strconv"
	"strings"
)

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

func WillRun(labels, notLabels, hostLabels, hostNotLabels map[string]bool) bool {
	// 1. Check every test label is in the hostLabels
	for l, _ := range labels {
		if _, ok := hostLabels[l]; !ok {
			return false
		}
	}
	// 2. Check every test notLabel is NOT in the hostLabels
	for l, _ := range notLabels {
		if _, ok := hostLabels[l]; ok {
			return false
		}
	}
	// 3. Check that none of the test labels appear in the hostNotLabels
	for l, _ := range hostNotLabels {
		if _, ok := labels[l]; ok {
			return false
		}
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
