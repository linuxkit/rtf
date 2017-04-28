package local

import (
	"os/exec"

	"github.com/dave-tucker/rtf/logger"
	"github.com/dave-tucker/rtf/sysinfo"

	"time"
)

const (
	GroupFile    = "group.sh"
	PreTestFile  = "pre-test.sh"
	PostTestFile = "post-test.sh"
	TestFile     = "test.sh"
)

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

type TestResult int

const (
	Pass = iota
	Fail
	Skip
	Cancel
)

type Result struct {
	TestResult TestResult
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Name       string
	Summary    string
	Labels     string
}

type OSInfo struct {
	OS      string
	Version string
	Name    string
	Arch    string
}

type RunConfig struct {
	Extra      bool
	CaseDir    string
	LogDir     string
	Logger     logger.LogDispatcher
	SystemInfo sysinfo.SystemInfo
	Labels     map[string]bool
	NotLabels  map[string]bool
}
