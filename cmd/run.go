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
	latestResults  = "latest"
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

// listCmd represents the list command
var runCmd = &cobra.Command{
	Use:   "run [test pattern]",
	Short: "A brief description of your command",
	RunE:  run,
}

func init() {
	RootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Expected only one test pattern")
	}

	pattern := strings.Join(args, "")

	systemInfo := sysinfo.GetSystemInfo()
	l, nl := local.ParseLabels(labels)
	for _, v := range systemInfo.List() {
		if _, ok := l[v]; !ok {
			l[v] = true
		}
	}

	var labelList []string
	for k := range l {
		labelList = append(labelList, k)
	}
	for k := range nl {
		labelList = append(labelList, "!"+k)
	}
	fmt.Printf("LABELS: %s\n", strings.Join(labelList, ", "))

	p, err := local.NewProject(caseDir)
	if err != nil {
		return err
	}
	if err := p.Init(); err != nil {
		return err
	}

	id := uuid.Generate()
	fmt.Printf("ID: %s\n", id)
	baseDir, err := setupResultsDirectory(id.String())
	if err != nil {
		return err
	}
	testsLogPath := filepath.Join(baseDir, testsLogName)
	testsCsvPath := filepath.Join(baseDir, testsCsvName)
	summaryCsvPath := filepath.Join(baseDir, summaryCsvName)

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
	runConfig := local.RunConfig{
		TestPattern: pattern,
		Extra:       extra,
		CaseDir:     caseDir,
		LogDir:      baseDir,
		Logger:      log,
		SystemInfo:  systemInfo,
		Labels:      l,
		NotLabels:   nl,
	}
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

func setupResultsDirectory(id string) (string, error) {
	baseDir, err := filepath.Abs(filepath.Join(resultDir, id))
	if err != nil {
		return "", err
	}

	_, err = os.Stat(baseDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return "", err
		}
	}

	linkPath := filepath.Join(resultDir, latestResults)
	_, err = os.Lstat(linkPath)
	if err == nil {
		if err := os.Remove(linkPath); err != nil {
			return "", err
		}
	}

	if err := os.Symlink(baseDir, linkPath); err != nil {
		return "", err
	}

	return baseDir, nil
}
