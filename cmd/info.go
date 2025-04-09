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
	"encoding/csv"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/linuxkit/rtf/local"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print test cases and their descriptions",
	RunE:  info,
}

var (
	csvInfo bool
)

func init() {
	flags := infoCmd.Flags()
	flags.BoolVarP(&csvInfo, "csv", "", false, "Generate a CSV file")
	RootCmd.AddCommand(infoCmd)
}

func info(_ *cobra.Command, _ []string) error {
	config := local.NewRunConfig(labels, "")
	p, err := local.InitNewProject(caseDir)
	if err != nil {
		return err
	}

	tw := new(tabwriter.Writer)
	tw.Init(os.Stdout, 0, 8, 0, '\t', 0)

	cw := csv.NewWriter(os.Stdout)

	lst := p.List(config)
	if !csvInfo {
		_, _ = fmt.Fprintf(tw, "NAME\tDESCRIPTION\n")
	} else {
		heading := []string{"Name", "Description", "Known issues"}
		if err := cw.Write(heading); err != nil {
			return nil
		}
	}

	for _, i := range lst {
		if !csvInfo {
			_, _ = fmt.Fprintf(tw, "%s\t%s\n", i.Name, i.Summary)
		} else {
			out := []string{i.Name, i.Summary, i.Issue}
			if err := cw.Write(out); err != nil {
				return nil
			}
		}
	}
	_ = tw.Flush()
	cw.Flush()
	return nil
}
