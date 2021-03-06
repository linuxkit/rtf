package local

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/linuxkit/rtf/logger"
)

// NewProject creates a new top-level Group at the provided path
func NewProject(path string) (*Group, error) {
	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
	}
	g := &Group{Parent: nil, Path: path}
	return g, nil
}

// InitNewProject creates a new Group, and calls Init() on it
func InitNewProject(path string) (*Group, error) {
	group, err := NewProject(path)
	if err != nil {
		return group, err
	}
	return group, group.Init()
}

// NewGroup creates a new Group with the given parent and path
func NewGroup(parent *Group, path string) (*Group, error) {
	g := &Group{Parent: parent, Path: path, PreTestPath: parent.PreTestPath, PostTestPath: parent.PostTestPath}
	if err := g.Init(); err != nil {
		return nil, err
	}
	return g, nil
}

// IsGroup determines if a path contains a group
func IsGroup(path string) bool {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
		return false
	}

	for _, file := range files {
		if file.Name() == GroupFileName+".sh" || file.Name() == GroupFileName+".ps1" {
			return true
		}
		if file.IsDir() {
			return true
		}
	}
	return false
}

// Init is the group initialization function and should be called immediately after a group has been created
func (g *Group) Init() error {
	g.GroupFilePath, _ = checkScript(g.Path, GroupFileName)

	tags, err := ParseTags(g.GroupFilePath)
	if err != nil {
		tags = &Tags{}
	}
	g.Tags = tags

	var name string
	var order int

	g.Labels, g.NotLabels = ParseLabels(g.Tags.Labels)

	order, name = getNameAndOrder(filepath.Base(g.Path))

	if g.Parent == nil {
		// top of tree
		g.PreTestPath, _ = checkScript(g.Path, PreTestFileName)
		g.PostTestPath, _ = checkScript(g.Path, PostTestFileName)
	} else {
		g.Tags.Name = fmt.Sprintf("%s.%s", g.Parent.Name(), name)
	}
	g.order = order

	files, err := ioutil.ReadDir(g.Path)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			if strings.HasPrefix(f.Name(), "_") {
				// ignore
				continue
			}
			path := filepath.Join(g.Path, f.Name())
			if IsGroup(path) {
				gc, err := NewGroup(g, path)
				if err != nil {
					return err
				}
				g.Children = append(g.Children, gc)
			}
			if IsTest(path) {
				t, err := NewTest(g, path)
				if err != nil {
					return err
				}
				g.Children = append(g.Children, t)
			}
		}
	}
	return nil
}

// LabelString provides all labels in a comma separated list
func (g *Group) LabelString() string {
	return makeLabelString(g.Labels, g.NotLabels, ", ")
}

// Name returns the name of the group
func (g *Group) Name() string {
	return g.Tags.Name
}

// List lists all child groups and tests
func (g *Group) List(config RunConfig) []Info {
	sort.Sort(ByOrder(g.Children))

	if !g.willRun(config) {
		return []Info{{
			TestResult: Skip,
			Name:       g.Name(),
			Labels:     g.Labels,
			NotLabels:  g.NotLabels,
		}}
	}

	infos := []Info{}
	for _, c := range g.Children {
		lst := c.List(config)
		infos = append(infos, lst...)
	}

	return infos
}

// Run will run all child groups and tests
func (g *Group) Run(config RunConfig) ([]Result, error) {
	var results []Result
	sort.Sort(ByOrder(g.Children))

	if !g.willRun(config) {
		return []Result{{TestResult: Skip,
			Name:      g.Name(),
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}}, nil
	}

	if g.GroupFilePath != "" {
		config.Logger.Log(logger.LevelInfo, fmt.Sprintf("%s::ginit()", g.Name()))
		res, err := executeScript(g.GroupFilePath, g.Path, "", []string{"init"}, config)
		if err != nil {
			return results, err
		}
		if res.TestResult != Pass {
			return results, fmt.Errorf("Error running %s", g.GroupFilePath+":init")
		}
	}

	if config.Parallel {
		var wg sync.WaitGroup
		resCh := make(chan []Result, len(g.Children))
		errCh := make(chan error, len(g.Children))

		for _, c := range g.Children {
			wg.Add(1)
			go func(c TestContainer, cf RunConfig) {
				defer wg.Done()
				res, err := c.Run(cf)
				if err != nil {
					errCh <- err
				}
				resCh <- res
			}(c, config)
		}

		go func() {
			wg.Wait()
			close(resCh)
			close(errCh)
		}()

		for err := range errCh {
			if err != nil {
				return results, err
			}
		}

		for res := range resCh {
			results = append(results, res...)
		}
	} else {
		for _, c := range g.Children {
			res, err := c.Run(config)
			if err != nil {
				return results, err
			}
			results = append(results, res...)
		}
	}

	if g.GroupFilePath != "" {
		config.Logger.Log(logger.LevelInfo, fmt.Sprintf("%s::gdeinit()", g.Name()))
		res, err := executeScript(g.GroupFilePath, g.Path, "", []string{"deinit"}, config)
		if err != nil {
			return results, err
		}
		if res.TestResult != Pass {
			return results, fmt.Errorf("Error running %s", g.GroupFilePath+":deinit")
		}
	}
	return results, nil
}

// Order returns the order of a group
func (g *Group) Order() int {
	return g.order
}

// willRun determines if tests from this group should be run based on labels and runtime config.
func (g *Group) willRun(config RunConfig) bool {
	if !CheckLabel(g.Labels, g.NotLabels, config) {
		return false
	}

	if config.TestPattern == "" {
		return true
	}

	return strings.HasPrefix(config.TestPattern, g.Name()) || strings.HasPrefix(g.Name(), config.TestPattern)
}
