package local

import (
	"fmt"

	"github.com/linuxkit/rtf/logger"
)

// Run satisfies the TestContainer interface.
// Run the group init or deinit command.
func (g GroupCommand) Run(config RunConfig) ([]Result, error) {
	config.Logger.Log(logger.LevelInfo, fmt.Sprintf("%s::%s()", g.Name, g.Type))
	res, err := executeScript(g.FilePath, g.Path, "", []string{g.Type}, config)
	if err != nil {
		return nil, err
	}
	if res.TestResult != Pass {
		return nil, fmt.Errorf("error running %s:%s", g.FilePath, g.Type)
	}
	return []Result{res}, nil
}

// List satisfies the TestContainer interface
func (g GroupCommand) List(config RunConfig) []Info {
	info := Info{
		Name: g.Name,
	}

	return []Info{info}
}

// Order returns a tests order
func (g GroupCommand) Order() int {
	if g.Type == "init" {
		return 0
	}
	return 1
}

// Gather satisfies the TestContainer interface
func (g GroupCommand) Gather(config RunConfig, count int) ([]TestContainer, int) {
	return []TestContainer{&g}, 0
}
