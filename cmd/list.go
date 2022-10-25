// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/linuxkit/rtf/local"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List test cases",
	RunE:  list,
}

func init() {
	flags := listCmd.Flags()
	// shardPattern is 1-based (1/10, 3/10, 10/10) rather than normal computer 0-based (0/9, 2/9, 9/9), because it is easier for
	// humans to understand when calling the CLI.
	flags.StringVarP(&shardPattern, "shard", "s", "", "which shard to run, in form of 'N/M' where N is the shard number and M is the total number of shards, smallest shard number is 1")
	RootCmd.AddCommand(listCmd)
}

func list(_ *cobra.Command, args []string) error {
	shard, totalShards, err := parseShardPattern(shardPattern)
	if err != nil {
		return err
	}
	pattern, err := local.ValidatePattern(args)
	if err != nil {
		return err
	}
	config := local.NewRunConfig(labels, pattern)

	p, err := local.InitNewProject(caseDir)
	if err != nil {
		return err
	}
	if totalShards > 0 {
		if err := p.SetShard(shard, totalShards); err != nil {
			return err
		}
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	lst := p.List(config)
	fmt.Fprint(w, "STATE\tTEST\tLABELS\n")
	for _, i := range lst {
		state := i.TestResult.Sprintf(local.TestResultNames[i.TestResult])
		fmt.Fprintf(w, "%s\t%s\t%s\n", state, i.Name, i.LabelString())
	}
	w.Flush()
	return nil
}
