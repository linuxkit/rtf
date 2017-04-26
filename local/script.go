package local

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dave-tucker/rtf/logger"
)

func executeScript(script, cwd, name, labels string, args []string, config RunConfig) (Result, error) {
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
	cmd := exec.Command("/bin/sh", cmdArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Result{}, err
	}

	osEnv := os.Environ()
	goPath := os.Getenv("GOPATH")
	libDir := filepath.Join(goPath, "src", "github.com", "dave-tucker", "rtf", "lib", "lib.sh")
	utilsDir := filepath.Join(goPath, "src", "github.com", "dave-tucker", "rtf", "bin")
	currentDir, err := os.Getwd()
	if err != nil {
		return Result{}, err
	}

	projectDir, err := filepath.Abs(config.CaseDir)
	if err != nil {
		return Result{}, err
	}

	envPath := os.Getenv("PATH")
	rtEnv := []string{
		fmt.Sprintf("RT_ROOT=%s", currentDir),
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
