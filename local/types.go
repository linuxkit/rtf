package local

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/linuxkit/rtf/logger"
	"github.com/linuxkit/rtf/sysinfo"
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

// TestResultNames provides a mapping of numerical result values to human readable strings
var TestResultNames = map[TestResult]string{
	Pass:   "Pass",
	Fail:   "Fail",
	Skip:   "Skip",
	Cancel: "Cancel",
}

// Sprintf prints the arguments using fmt.Sprintf but colours it depending on the TestResult
func (r TestResult) Sprintf(format string, a ...interface{}) string {
	switch r {
	case Pass:
		return color.GreenString(format, a...)
	case Fail:
		return color.RedString(format, a...)
	case Cancel:
		return color.YellowString(format, a...)
	case Skip:
		return color.YellowString(format, a...)
	}
	return fmt.Sprintf(format, a...)
}

// TestResultColorFunc provides a mapping of numerical result values to a fmt.Sprintf() style function
var TestResultColorFunc = map[TestResult]func(a ...interface{}) string{}

// Result encapsulates a TestResult and additional data about a test run
type Result struct {
	Test            *Test         `json:"-"`
	Name            string        `json:"name,omitempty"` // Name may be different to Test.Name() for repeated tests.
	TestResult      TestResult    `json:"result"`
	BenchmarkResult string        `json:"benchmark,omitempty"`
	StartTime       time.Time     `json:"start,omitempty"`
	EndTime         time.Time     `json:"end,omitempty"`
	Duration        time.Duration `json:"duration,omitempty"`
}

// Info encapsulates the information necessary to list tests and test groups
type Info struct {
	Name       string
	TestResult TestResult
	Summary    string
	Issue      string
	Labels     map[string]bool
	NotLabels  map[string]bool
}

// LabelString returns all labels in a comma separated string
func (i *Info) LabelString() string {
	return makeLabelString(i.Labels, i.NotLabels, ", ")
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
	List(config RunConfig) []Info
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

// Summary contains a summary of a whole run, mostly used for writing out a JSON file
type Summary struct {
	ID         string             `json:"id,omitempty"`
	StartTime  time.Time          `json:"start,omitempty"`
	EndTime    time.Time          `json:"end,omitempty"`
	SystemInfo sysinfo.SystemInfo `json:"system,omitempty"`
	Labels     []string           `json:"labels,omitempty"`
	Results    []Result           `json:"results,omitempty"`
}
