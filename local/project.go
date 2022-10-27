package local

import (
	"errors"
	"fmt"
	"path/filepath"
)

// NewProject creates a new top-level Group at the provided path
func NewProject(path string) (*Project, error) {
	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
	}
	g := &Project{&Group{Parent: nil, Path: path}, 0, 0}
	return g, nil
}

// InitNewProject creates a new Group, and calls Init() on it
func InitNewProject(path string) (*Project, error) {
	project, err := NewProject(path)
	if err != nil {
		return project, err
	}
	return project, project.Init()
}

// SetShard returns a subset of the group based on the shards
func (p *Project) SetShard(shard, totalShards int) error {
	if shard < 1 || shard > totalShards {
		return errors.New("shard must be between 1 and totalShards")
	}
	if totalShards < 1 {
		return fmt.Errorf("totalShards must be greater than 0")
	}
	if totalShards == 1 {
		return nil
	}
	p.shard = shard
	p.totalShards = totalShards

	return nil
}

// Run runs all child groups and tests, limited by the provided shards
func (p *Project) Run(config RunConfig) ([]Result, error) {
	// if we are sharded, walk the tree to create a list of the tests to run
	if p.totalShards <= 1 {
		return p.Group.Run(config)
	}

	infos := p.Group.List(config)
	start, count := calculateShard(len(infos), p.shard, p.totalShards)

	config.start = start
	config.count = count

	return p.Group.Run(config)
}

// List lists all child groups and tests, limited by the provided shards
func (p *Project) List(config RunConfig) []Info {
	infos := p.Group.List(config)
	if p.totalShards <= 1 {
		return infos
	}
	start, count := calculateShard(len(infos), p.shard, p.totalShards)
	return infos[start : start+count]
}
