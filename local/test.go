package local

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/linuxkit/rtf/logger"
)

// NewTest creates a new test
func NewTest(group *Group, path string) (*Test, error) {
	t := &Test{Parent: group, Path: path}
	if err := t.Init(); err != nil {
		return nil, err
	}
	return t, nil
}

// IsTest determines if a path contains a test or not
func IsTest(path string) bool {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
		return false
	}

	for _, file := range files {
		if file.Name() == TestFile {
			return true
		}
	}
	return false
}

// Init initializes a test and should be run immmediately after NewTest
func (t *Test) Init() error {
	tf := filepath.Join(t.Path, TestFile)
	tags, err := ParseTags(tf)
	if err != nil {
		return err
	}
	t.Tags = tags
	t.Summary = tags.Summary
	order, name := getNameAndOrder(filepath.Base(t.Path))
	if t.Parent == nil {
		return fmt.Errorf("A test should have a parent group")
	}
	t.Tags.Name = fmt.Sprintf("%s.%s", t.Parent.Name(), name)
	t.Labels, t.NotLabels = ParseLabels(t.Tags.Labels)
	for k, v := range t.Parent.Labels {
		if ok := t.Labels[k]; !ok {
			t.Labels[k] = v
		}
	}
	for k, v := range t.Parent.NotLabels {
		if ok := t.NotLabels[k]; !ok {
			t.NotLabels[k] = v
		}
	}
	t.order = order
	return nil
}

// Name returns the test's name
func (t *Test) Name() string {
	return t.Tags.Name
}

// LabelString returns all labels in a comma separated string
func (t *Test) LabelString() string {
	return makeLabelString(t.Labels, t.NotLabels)
}

// List satisfies the TestContainer interface
func (t *Test) List(config RunConfig) []Result {
	if !t.willRun(config) {
		return []Result{{
			TestResult: Skip,
			Name:       t.Name(),
			Summary:    t.Tags.Summary,
			Labels:     t.LabelString(),
		}}
	}
	return []Result{{
		Name:    t.Name(),
		Summary: t.Tags.Summary,
		Labels:  t.LabelString(),
	}}
}

// Run runs a test
func (t *Test) Run(config RunConfig) ([]Result, error) {
	var results []Result
	appendIteration := false

	if !t.willRun(config) {
		config.Logger.Log(logger.LevelSkip, fmt.Sprintf("%s %.2fs", t.Name(), 0.0))
		return []Result{{TestResult: Skip,
			Name: t.Name(),
		}}, nil
	}

	if t.Tags.Repeat == 0 {
		// Always run at least once
		t.Tags.Repeat = 1
	} else {
		appendIteration = true
	}

	for i := 1; i < t.Tags.Repeat+1; i++ {
		name := t.Name()
		if appendIteration {
			name = fmt.Sprintf("%s.%d", name, i)
		}

		logFileName := filepath.Join(config.LogDir, name+".log")
		logFile, err := os.Create(logFileName)
		if err != nil {
			return nil, err
		}
		testLogger := logger.NewFileLogger(logFile)
		testLogger.SetLevel(logger.LevelDebug)
		config.Logger.Register(logFileName, testLogger)
		defer config.Logger.Unregister(logFileName)

		if t.Parent.PreTest != "" {
			res, err := executeScript(t.Parent.PreTest, t.Path, name, t.LabelString(), []string{name}, config)
			if res.TestResult != Pass {
				return results, fmt.Errorf("Error running: %s. %s", t.Parent.PreTest, err.Error())
			}
		}
		// Run the test
		config.Logger.Log(logger.LevelInfo, fmt.Sprintf("Running Test %s in %s", name, t.Path))
		tf := filepath.Join(t.Path, TestFile)
		res, err := executeScript(tf, t.Path, name, t.LabelString(), nil, config)
		if err != nil {
			return results, err
		}
		switch res.TestResult {
		case Pass:
			config.Logger.Log(logger.LevelPass, fmt.Sprintf("%s %.2fs", res.Name, res.Duration.Seconds()))
		case Fail:
			config.Logger.Log(logger.LevelFail, fmt.Sprintf("%s %.2fs", res.Name, res.Duration.Seconds()))
		case Cancel:
			config.Logger.Log(logger.LevelCancel, fmt.Sprintf("%s %.2fs", res.Name, res.Duration.Seconds()))
		}
		if t.Parent.PostTest != "" {
			res, err := executeScript(t.Parent.PostTest, t.Path, name, t.LabelString(), []string{name, fmt.Sprintf("%d", res.TestResult)}, config)
			if res.TestResult != Pass {
				return results, fmt.Errorf("Error running: %s. %s", t.Parent.PreTest, err.Error())
			}
		}
		results = append(results, res)
	}
	return results, nil
}

// Order returns a tests order
func (t *Test) Order() int {
	return t.order
}

// willRun determines if the test should be run based on labels and runtime config.
func (t *Test) willRun(config RunConfig) bool {
	if !CheckLabel(t.Labels, t.NotLabels, config) {
		return false
	}

	// HasPrefix matches on "" for config.TestPattern
	return strings.HasPrefix(t.Name(), config.TestPattern)
}
