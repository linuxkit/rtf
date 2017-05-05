package local

import (
	"os/exec"

	"github.com/linuxkit/rtf/logger"
	"github.com/linuxkit/rtf/sysinfo"

	"time"
)

const (
	// GroupFile is the name of the group script
	GroupFile = "group.sh"
	// PreTestFile is the name of a pre-test script
	PreTestFile = "pre-test.sh"
	// PostTestFile is the name of a ppst-test script
	PostTestFile = "post-test.sh"
	// TestFile is the name of a test script
	TestFile = "test.sh"
)

// Group is a group of tests and other groups
type Group struct {
	Tags      *Tags
	PreTest   string
	PostTest  string
	Parent    *Group
	Order     int
	Path      string
	Labels    map[string]bool
	NotLabels map[string]bool
	Tests     []*Test
	Children  []*Group
}

// Test is a test
type Test struct {
	Parent    *Group
	Tags      *Tags
	Path      string
	Command   exec.Cmd
	Repeat    int
	Order     int
	Summary   string
	Author    string
	Labels    map[string]bool
	NotLabels map[string]bool
}

// TestResult is the result of a test run
type TestResult int

const (
	// Pass is a test pass
	Pass = iota
	// Fail is a test failure
	Fail
	// Skip is a test skip
	Skip
	// Cancel is a test cancellation
	Cancel
)

// Result encapsulates a TestResult and additional data about a test run
type Result struct {
	TestResult TestResult
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Name       string
	Summary    string
	Labels     string
}

// OSInfo contains information about the OS the tests are running on
type OSInfo struct {
	OS      string
	Version string
	Name    string
	Arch    string
}

// RunConfig contains runtime configuration information
type RunConfig struct {
	Extra       bool
	CaseDir     string
	LogDir      string
	Logger      logger.LogDispatcher
	SystemInfo  sysinfo.SystemInfo
	Labels      map[string]bool
	NotLabels   map[string]bool
	TestPattern string
}
