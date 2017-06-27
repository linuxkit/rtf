package local

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// NewGroup creates a new Group with the given parent and path
func NewGroup(parent *Group, path string) (*Group, error) {
	g := &Group{Parent: parent, Path: path, PreTest: parent.PreTest, PostTest: parent.PostTest}
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
		if file.Name() == GroupFile {
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
	gf := filepath.Join(g.Path, GroupFile)
	tags, err := ParseTags(gf)
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
		pre := filepath.Join(g.Path, PreTestFile)
		post := filepath.Join(g.Path, PostTestFile)
		if _, err := os.Stat(pre); err == nil {
			g.PreTest = pre
		}

		if _, err := os.Stat(post); err == nil {
			g.PostTest = post
		}

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
	return makeLabelString(g.Labels, g.NotLabels)
}

// Name returns the name of the group
func (g *Group) Name() string {
	return g.Tags.Name
}

// List lists all child groups and tests
func (g *Group) List(config RunConfig) []Result {
	result := []Result{}
	sort.Sort(ByOrder(g.Children))

	if !WillRun(g.Labels, g.NotLabels, config) {
		return []Result{{
			TestResult: Skip,
			Name:       g.Name(),
			Summary:    g.Tags.Summary,
			Labels:     g.LabelString(),
		}}
	}

	for _, c := range g.Children {
		lst := c.List(config)
		result = append(result, lst...)
	}

	return result
}

// Run will run all child groups and tests
func (g *Group) Run(config RunConfig) ([]Result, error) {
	var results []Result
	sort.Sort(ByOrder(g.Children))

	if !WillRun(g.Labels, g.NotLabels, config) {
		return []Result{{TestResult: Skip,
			Name:      g.Name(),
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}}, nil
	}

	init := false
	gfName := filepath.Join(g.Path, GroupFile)
	_, err := os.Stat(gfName)
	if err != nil {
		if !os.IsNotExist(err) {
			return results, err
		}
	} else {
		init = true
	}

	if init {
		config.Logger.Log(logger.LevelInfo, fmt.Sprintf("%s::ginit()", g.Name()))
		res, err := executeScript(gfName, g.Path, "", g.LabelString(), []string{"init"}, config)
		if err != nil {
			return results, err
		}
		if res.TestResult != Pass {
			return results, fmt.Errorf("Error running %s", gfName+":init")
		}
	}

	for _, c := range g.Children {
		res, err := c.Run(config)
		if err != nil {
			return results, err
		}
		results = append(results, res...)
	}

	if init {
		config.Logger.Log(logger.LevelInfo, fmt.Sprintf("%s::gdeinit()", g.Name()))
		res, err := executeScript(gfName, g.Path, "", g.LabelString(), []string{"deinit"}, config)
		if err != nil {
			return results, err
		}
		if res.TestResult != Pass {
			return results, fmt.Errorf("Error running %s", gfName+":deinit")
		}
	}
	return results, nil
}

// Order returns the order of a group
func (g *Group) Order() int {
	return g.order
}
