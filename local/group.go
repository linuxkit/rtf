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
)

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

// Gather gathers all runnable child groups and tests
func (g *Group) Gather(config RunConfig) ([]TestContainer, int) {
	sort.Sort(ByOrder(g.Children))

	if !g.willRun(config) {
		return nil, 0
	}
	containers := []TestContainer{}
	var subCount int

	if g.GroupFilePath != "" {
		containers = append(containers, GroupCommand{Name: g.Name(), FilePath: g.GroupFilePath, Path: g.Path, Type: "init"})
	}

	for _, c := range g.Children {
		lst, childCount := c.Gather(config)
		// if we had no runnable tests, do not bother adding the group init/deinit, just return the empty list
		if childCount == 0 {
			continue
		}
		containers = append(containers, lst...)
		subCount += childCount
	}

	if g.GroupFilePath != "" {
		containers = append(containers, GroupCommand{Name: g.Name(), FilePath: g.GroupFilePath, Path: g.Path, Type: "deinit"})
	}

	return containers, subCount
}

// Run will run all child groups and tests
func (g *Group) Run(config RunConfig) ([]Result, error) {
	var results []Result

	// This gathers all of the individual tests and group init/deinit commands
	// all the way down, leading to a flat list we can execute, rather than recursion.
	// That should make it easier to break into shards.
	runnables, _ := g.Gather(config)
	if len(runnables) == 0 {
		return []Result{{TestResult: Skip,
			Name:      g.Name(),
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}}, nil
	}

	if config.Parallel {
		var wg sync.WaitGroup
		resCh := make(chan []Result, len(g.Children))
		errCh := make(chan error, len(g.Children))

		for _, c := range runnables {
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
		for _, r := range runnables {
			res, err := r.Run(config)
			if err != nil {
				return results, err
			}
			results = append(results, res...)
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

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// calculateShard calculate the start and end of a slice to use in a given shard of a given total
// slice and a given number of shards.
// We split by the total list size. If it is uneven, e.g. 22 elements into 10 shards,
// then the first shards will be rounded up until the reminder can be split evenly.
//
// e.g.
// 22 elements into 10 shards will be 3, 3, 2, 2, 2, 2, 2, 2, 2, 2
// 29 elements into 10 shards will be 3, 3, 3, 3, 3, 3, 3, 3, 3, 2
// 30 elements into 10 shards will be 3, 3, 3, 3, 3, 3, 3, 3, 3, 3
// 8 elements into 5 shards will be 2, 2, 2, 1, 1
//
// Ths important thing is consistency among runs using the same set of parameters
// so that you can reliably get the same subset each time.
func calculateShard(size, shard, totalShards int) (start, count int) {
	elmsPerShard := size / totalShards
	remainder := size % totalShards
	before := (shard - 1) * elmsPerShard
	if remainder > 0 {
		before += min(remainder, shard-1)
	}
	count = elmsPerShard
	if remainder >= shard {
		count++
	}
	return before, count
}
