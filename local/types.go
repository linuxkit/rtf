package local

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/linuxkit/rtf/logger"
	"github.com/linuxkit/rtf/sysinfo"

	"time"
)

const (
	// GroupFileName is the name of the group script (without the extension)
	GroupFileName = "group"
	// PreTestFileName is the name of a pre-test script (without the extension)
	PreTestFileName = "pre-test"
	// PostTestFileName is the name of a post-test script (without the extension)
	PostTestFileName = "post-test"
	// TestFileName is the name of a test script (without the extension)
	TestFileName = "test"
)

// checkScript checks if a script with 'name' exists in 'path'
func checkScript(path, name string) (string, error) {
	// On Windows, powershell scripts take precedence.
	if runtime.GOOS == "windows" {
		f := filepath.Join(path, name+".ps1")
		if _, err := os.Stat(f); err == nil {
			return f, nil
		}
	}

	f := filepath.Join(path, name+".sh")
	if _, err := os.Stat(f); err != nil {
		// On non-windows, shell scripts take precedence but we check for powershell too
		if runtime.GOOS != "windows" {
			f := filepath.Join(path, name+".ps1")
			if _, err := os.Stat(f); err == nil {
				return f, nil
			}
		}
		return "", err
	}
	return f, nil
}

// Group is a group of tests and other groups
type Group struct {
	Parent        *Group
	Tags          *Tags
	Path          string
	GroupFilePath string
	PreTestPath   string
	PostTestPath  string
	order         int
	Labels        map[string]bool
	NotLabels     map[string]bool
	Children      []TestContainer
}

// Test is a test
type Test struct {
	Parent       *Group
	Tags         *Tags
	Path         string
	TestFilePath string
	Command      exec.Cmd
	Repeat       int
	order        int
	Summary      string
	Author       string
	Labels       map[string]bool
	NotLabels    map[string]bool
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
	TestResult      TestResult
	BenchmarkResult string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Name            string
	Summary         string
	Labels          string
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

// A TestContainer is a container that can hold one or more tests
type TestContainer interface {
	Order() int
	List(config RunConfig) []Result
	Run(config RunConfig) ([]Result, error)
}

// ByOrder implements the sort.Sorter interface for TestContainer
type ByOrder []TestContainer

// Len returns the length of the []TestContainer
func (a ByOrder) Len() int { return len(a) }

// Swap swaps two items in a []TestContainer
func (a ByOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less compares whether the order of i is less than that of j
func (a ByOrder) Less(i, j int) bool { return a[i].Order() < a[j].Order() }
