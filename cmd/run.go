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
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	summaryJSONName = "SUMMARY.json"
	testsCsvName    = "TESTS.csv"
	summaryCsvName  = "SUMMARY.csv"
	testsLogName    = "TESTS.log"
	latestResults   = "latest"
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
		"System Model",
		"CPU",
		"Memory",
	}
	testCsvFields = []string{
		"ID",
		"Timestamp",
		"Duration",
		"Name",
		"Result",
		"Benchmark",
		"Description",
		"Issues",
	}
)

var (
	resultDir    string
	id           string
	symlink      bool
	extra        bool
	parallel     bool
	shardPattern string
)

var runCmd = &cobra.Command{
	Use:   "run [test pattern]",
	Short: "Run test cases",
	RunE:  run,
}

func init() {
	flags := runCmd.Flags()
	flags.StringVarP(&resultDir, "resultdir", "r", "_results", "Directory to place results in")
	flags.StringVarP(&id, "id", "", "", "ID for this test run")
	flags.BoolVarP(&extra, "extra", "x", false, "Add extra debug info to log files")
	flags.BoolVarP(&parallel, "parallel", "p", false, "Run multiple tests in parallel")
	// shardPattern is 1-based (1/10, 3/10, 10/10) rather than normal computer 0-based (0/9, 2/9, 9/9), because it is easier for
	// humans to understand when calling the CLI.
	flags.StringVarP(&shardPattern, "shard", "s", "", "which shard to run, in form of 'N/M' where N is the shard number and M is the total number of shards, smallest shard number is 1. Shards are applied only to tests that would run, not those that would be skipped.")
	RootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) error {
	shard, totalShards, err := parseShardPattern(shardPattern)
	if err != nil {
		return err
	}
	pattern, err := local.ValidatePattern(args)
	if err != nil {
		return err
	}
	runConfig := local.NewRunConfig(labels, pattern)
	runConfig.Extra = extra
	runConfig.Parallel = parallel

	p, err := local.InitNewProject(caseDir)
	if err != nil {
		return err
	}
	if totalShards > 0 {
		if err := p.SetShard(shard, totalShards); err != nil {
			return err
		}
	}

	var labelList []string
	for k := range runConfig.Labels {
		labelList = append(labelList, k)
	}
	for k := range runConfig.NotLabels {
		labelList = append(labelList, "!"+k)
	}
	fmt.Printf("LABELS: %s\n", strings.Join(labelList, ", "))

	if id == "" {
		symlink = true
		id = uuid.Generate().String()
	}

	fmt.Printf("ID: %s\n", id)
	baseDir, err := setupResultsDirectory(id, symlink)
	if err != nil {
		return err
	}
	testsLogPath := filepath.Join(baseDir, testsLogName)
	testsCsvPath := filepath.Join(baseDir, testsCsvName)
	summaryCsvPath := filepath.Join(baseDir, summaryCsvName)
	summaryJSONPath := filepath.Join(baseDir, summaryJSONName)

	tf, err := os.Create(testsCsvPath)
	if err != nil {
		return err
	}
	defer func() { _ = tf.Close() }()

	tCsv := csv.NewWriter(tf)
	if err = tCsv.Write(testCsvFields); err != nil {
		return err
	}

	sf, err := os.Create(summaryCsvPath)
	if err != nil {
		return err
	}
	defer func() { _ = sf.Close() }()

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

	summary := local.Summary{
		ID:         id,
		SystemInfo: systemInfo,
		Labels:     labelList,
		StartTime:  startTime,
	}

	res, err := p.Run(runConfig)
	if err != nil {
		return err
	}

	for _, r := range res {
		if r.Test != nil {
			summary.Results = append(summary.Results, r)
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
		var testSummary, issue string
		if r.Test != nil {
			// Skipped test groups are in the result list but do not contain a Test reference
			testSummary = r.Test.Tags.Summary
			issue = r.Test.Tags.Issue
		}
		testResult := []string{
			id,
			r.EndTime.Format(time.RFC3339),
			strconv.FormatFloat(r.Duration.Seconds(), 'f', -1, 32),
			r.Name,
			local.TestResultNames[r.TestResult],
			r.BenchmarkResult,
			testSummary,
			issue,
		}
		if err = tCsv.Write(testResult); err != nil {
			return err
		}
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	summary.EndTime = endTime
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(summaryJSONPath, summaryJSON, 0644); err != nil {
		return err
	}

	summaryCSV := []string{
		id,
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
		systemInfo.Model,
		systemInfo.CPU,
		strconv.FormatInt(systemInfo.Memory, 10),
	}
	if err = sCsv.Write(summaryCSV); err != nil {
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
		return fmt.Errorf("some tests failed")
	}
	return nil
}

func setupResultsDirectory(id string, link bool) (string, error) {
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

	if link {
		linkPath := filepath.Join(resultDir, latestResults)
		_, err = os.Lstat(linkPath)
		if err == nil {
			if err := os.Remove(linkPath); err != nil {
				return "", err
			}
		}

		if err := os.Symlink(id, linkPath); err != nil {
			return "", err
		}
	}

	return baseDir, nil
}

func parseShardPattern(pattern string) (shard int, total int, err error) {
	if pattern == "" {
		return 0, 0, nil
	}
	parts := strings.SplitN(pattern, "/", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid shard pattern: %s", pattern)
	}
	if shard, err = strconv.Atoi(parts[0]); err != nil {
		return 0, 0, fmt.Errorf("invalid shard pattern: %s", pattern)
	}
	if total, err = strconv.Atoi(parts[1]); err != nil {
		return 0, 0, fmt.Errorf("invalid shard pattern: %s", pattern)
	}
	if shard < 1 || total < 1 || shard > total {
		return 0, 0, fmt.Errorf("invalid shard pattern: %s", pattern)
	}
	return shard, total, nil
}
