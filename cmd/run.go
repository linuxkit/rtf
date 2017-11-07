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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/distribution/uuid"
	"github.com/linuxkit/rtf/local"
	"github.com/linuxkit/rtf/logger"
	"github.com/linuxkit/rtf/sysinfo"
	"github.com/spf13/cobra"
)

const (
	testsCsvName   = "TESTS.CSV"
	summaryCsvName = "SUMMARY.CSV"
	testsLogName   = "TESTS.log"
)

var (
	summaryCsvFields = []string{
		"ID",
		"Version",
		"Start Time",
		"End Time",
		"Duration",
		"Passed",
		"Failed",
		"Skipped",
		"Labels",
		"OS",
		"OS Name",
		"OS Version",
	}
	testCsvFields = []string{
		"ID",
		"Timestamp",
		"Duration",
		"Name",
		"Result",
		"Message",
	}
)

var (
	resultDir string
	extra     bool
)

var runCmd = &cobra.Command{
	Use:   "run [test pattern]",
	Short: "Run test cases",
	RunE:  run,
}

func init() {
	flags := runCmd.LocalFlags()
	flags.StringVarP(&resultDir, "resultdir", "r", "_results", "Directory to place results in")
	flags.BoolVarP(&extra, "extra", "x", false, "Add extra debug info to log files")
	RootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) error {
	pattern, err := local.ValidatePattern(args)
	if err != nil {
		return err
	}
	runConfig := local.NewRunConfig(labels, pattern)
	runConfig.Extra = extra

	p, err := local.InitNewProject(caseDir)
	if err != nil {
		return err
	}

	var labelList []string
	for k := range runConfig.Labels {
		labelList = append(labelList, k)
	}
	for k := range runConfig.NotLabels {
		labelList = append(labelList, "!"+k)
	}
	fmt.Printf("LABELS: %s\n", strings.Join(labelList, ", "))

	id := uuid.Generate()
	fmt.Printf("ID: %s\n", id)
	baseDir, err := filepath.Abs(filepath.Join(resultDir, id.String()))
	if err != nil {
		return err
	}
	testsLogPath := filepath.Join(baseDir, testsLogName)
	testsCsvPath := filepath.Join(baseDir, testsCsvName)
	summaryCsvPath := filepath.Join(baseDir, summaryCsvName)

	_, err = os.Stat(baseDir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(baseDir, 0755); err != nil {
			return err
		}
	}

	tf, err := os.Create(testsCsvPath)
	if err != nil {
		return err
	}
	defer tf.Close()

	tCsv := csv.NewWriter(tf)
	if err = tCsv.Write(testCsvFields); err != nil {
		return err
	}

	sf, err := os.Create(summaryCsvPath)
	if err != nil {
		return err
	}
	defer sf.Close()

	sCsv := csv.NewWriter(sf)

	if err = sCsv.Write(summaryCsvFields); err != nil {
		return err
	}

	lf, err := os.Create(testsLogPath)
	if err != nil {
		return err
	}

	testsLogger := logger.NewFileLogger(lf)
	consoleLogger := logger.NewConsoleLogger(true, nil)

	switch verbose {
	case 1:
		consoleLogger.SetLevel(logger.LevelStderr)
	case 2:
		consoleLogger.SetLevel(logger.LevelInfo)
	case 3:
		consoleLogger.SetLevel(logger.LevelDebug)
	default:
		consoleLogger.SetLevel(logger.LevelSummary)
	}
	testsLogger.SetLevel(logger.LevelDebug)
	log := logger.NewLogDispatcher(map[string]logger.Logger{testsLogName: testsLogger, "Console": consoleLogger})

	var passed, failed, skipped, cancelled int
	startTime := time.Now()
	runConfig.Logger = log
	runConfig.LogDir = baseDir
	runConfig.CaseDir = caseDir
	systemInfo := sysinfo.GetSystemInfo()
	runConfig.SystemInfo = systemInfo

	res, err := p.Run(runConfig)
	if err != nil {
		return err
	}

	for _, r := range res {
		testResult := []string{
			id.String(),
			r.EndTime.Format(time.RFC3339),
			strconv.FormatFloat(r.Duration.Seconds(), 'f', -1, 32),
			r.Name,
			fmt.Sprintf("%d", r.TestResult),
			"",
		}
		if err = tCsv.Write(testResult); err != nil {
			return err
		}
		switch r.TestResult {
		case local.Pass:
			passed++
		case local.Fail:
			failed++
		case local.Skip:
			skipped++
		case local.Cancel:
			cancelled++
		}
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	summary := []string{
		id.String(),
		"UNKNOWN",
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
		strconv.FormatFloat(duration.Seconds(), 'f', -1, 32),
		strconv.Itoa(passed),
		strconv.Itoa(failed),
		strconv.Itoa(skipped),
		"",
		systemInfo.OS,
		systemInfo.Name,
		systemInfo.Version,
	}
	if err = sCsv.Write(summary); err != nil {
		return err
	}

	tCsv.Flush()
	sCsv.Flush()

	log.Log(logger.LevelSummary, fmt.Sprintf("LogDir: %s", id))
	log.Log(logger.LevelSummary, fmt.Sprintf("Version: %s", systemInfo.Version))
	log.Log(logger.LevelSummary, fmt.Sprintf("Passed: %d", passed))
	log.Log(logger.LevelSummary, fmt.Sprintf("Failed: %d", failed))
	log.Log(logger.LevelSummary, fmt.Sprintf("Cancelled: %d", cancelled))
	log.Log(logger.LevelSummary, fmt.Sprintf("Skipped: %d", skipped))
	log.Log(logger.LevelSummary, fmt.Sprintf("Duration: %.2fs", duration.Seconds()))

	if failed > 0 {
		return fmt.Errorf("Some tests failed")
	}
	return nil
}
