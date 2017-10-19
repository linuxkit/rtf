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

	"github.com/fatih/color"
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
	RootCmd.AddCommand(listCmd)
}

func list(_ *cobra.Command, args []string) error {
	// FIXME: Colors appear to confuse the TabWriter
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	pattern, err := local.ValidatePattern(args)
	if err != nil {
		return err
	}
	config := local.NewRunConfig(labels, pattern)

	p, err := local.InitNewProject(caseDir)
	if err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	lst := p.List(config)
	fmt.Fprint(w, "STATE\tTEST\tLABELS\n")
	for _, t := range lst {
		var state string
		if t.TestResult == local.Skip {
			state = yellow("SKIP")
		} else {
			state = green("RUN")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", state, t.Name, t.Labels)
	}
	w.Flush()
	return nil
}
