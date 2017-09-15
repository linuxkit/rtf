package local

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/linuxkit/rtf/logger"
)

var shExecutable = "/bin/sh"

func init() {
	if runtime.GOOS != "windows" {
		return
	}

	// On Windows we need to find bash.exe and *not* pick up the one from WSL
	pathEnv := os.Getenv("PATH")
	for _, p := range strings.Split(pathEnv, ";") {
		t := filepath.Join(p, "bash.exe")
		if strings.Contains(t, "System32") || strings.Contains(t, "system32") {
			// This is where the WSL bash.exe lives
			continue
		}
		if _, err := os.Stat(t); err == nil {
			shExecutable = t
			break
		}
	}
}

func executeScript(script, cwd, name string, args []string, config RunConfig) (Result, error) {
	if name == "" {
		name = "UNKNOWN"
	}
	startTime := time.Now()
	var cmdArgs []string
	if config.Extra {
		cmdArgs = append(cmdArgs, "-x")

	}
	cmdArgs = append(cmdArgs, script)
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(shExecutable, cmdArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Result{}, err
	}

	osEnv := os.Environ()

	rootDir := os.Getenv("RT_ROOT")
	if rootDir == "" {
		// Assume the source is in the GOPATH
		goPath := os.Getenv("GOPATH")
		rootDir = filepath.Join(goPath, "src", "github.com", "linuxkit", "rtf")
	}
	libDir := filepath.Join(rootDir, "lib", "lib.sh")
	utilsDir := filepath.Join(rootDir, "bin")

	projectDir, err := filepath.Abs(config.CaseDir)
	if err != nil {
		return Result{}, err
	}

	labels := makeLabelString(config.Labels, config.NotLabels, ":")

	envPath := os.Getenv("PATH")
	rtEnv := []string{
		fmt.Sprintf("RT_ROOT=%s", rootDir),
		fmt.Sprintf("RT_UTILS=%s", utilsDir),
		fmt.Sprintf("RT_PROJECT_ROOT=%s", projectDir),
		fmt.Sprintf("RT_OS=%s", config.SystemInfo.OS),
		fmt.Sprintf("RT_OS_VER=%s", config.SystemInfo.Version),
		fmt.Sprintf("RT_LABELS=%s", labels),
		fmt.Sprintf("RT_TEST_NAME=%s", name),
		fmt.Sprintf("RT_LIB=%s", libDir),
		fmt.Sprintf("RT_RESULTS=%s", config.LogDir),
		fmt.Sprintf("PATH=%s:%s", utilsDir, envPath),
	}
	cmd.Env = append(osEnv, rtEnv...)
	if runtime.GOOS == "windows" {
		cmd.Env = append(cmd.Env, []string{"MSYS_NO_PATHCONV=1"}...)
	}

	cmd.Dir = cwd

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			config.Logger.Log(logger.LevelStdout, scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			config.Logger.Log(logger.LevelStderr, scanner.Text())
		}
	}()

	config.Logger.Log(logger.LevelInfo, fmt.Sprintf("Running command: %+v", cmd.Args))
	var res TestResult
	if err := cmd.Start(); err != nil {
		config.Logger.Log(logger.LevelCritical, err.Error())
		res = Fail
	}

	if res != Fail {
		if err := cmd.Wait(); err != nil {
			v, ok := err.(*exec.ExitError)
			if !ok {
				config.Logger.Log(logger.LevelCritical, err.Error())
				res = Fail
			} else {
				// FIXME: UNIX ONLY
				rc := v.Sys().(syscall.WaitStatus).ExitStatus()
				switch rc {
				case 253:
					res = Cancel
				default:
					res = Fail
				}
			}
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	return Result{
		TestResult: res,
		StartTime:  startTime,
		Duration:   duration,
		EndTime:    endTime,
		Name:       name,
	}, nil
}
