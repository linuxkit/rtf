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

var (
	shExecutable = "/bin/sh"
	psExecutable = "powershell.exe"
)

func init() {
	if runtime.GOOS != "windows" {
		t, err := exec.LookPath("pwsh")
		if err != nil {
			psExecutable = ""
		}
		psExecutable = t
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
	executable := shExecutable
	if filepath.Ext(script) == ".ps1" {
		executable = psExecutable
		cmdArgs = append(cmdArgs, []string{"-NoProfile", "-NonInteractive"}...)
	} else {
		if config.Extra {
			cmdArgs = append(cmdArgs, "-x")
		}
	}
	if executable == "" {
		return Result{}, fmt.Errorf("Can't find a suitable shell to execute %s", script)
	}
	cmdArgs = append(cmdArgs, script)
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(executable, cmdArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Result{}, err
	}

	rootDir := os.Getenv("RT_ROOT")
	if rootDir == "" {
		// Assume the source is in the GOPATH
		goPath := os.Getenv("GOPATH")
		rootDir = filepath.Join(goPath, "src", "github.com", "linuxkit", "rtf")
	}
	libDir := filepath.Join(rootDir, "lib", "lib.sh")
	if executable == psExecutable {
		libDir = filepath.Join(rootDir, "lib", "lib.ps1")
	}
	utilsDir := filepath.Join(rootDir, "bin")

	projectDir, err := filepath.Abs(config.CaseDir)
	if err != nil {
		return Result{}, err
	}

	labels := makeLabelString(config.Labels, config.NotLabels, ":")

	env := os.Environ()
	setEnv(&env, "RT_ROOT", rootDir)
	setEnv(&env, "RT_UTILS", utilsDir)
	setEnv(&env, "RT_PROJECT_ROOT", projectDir)
	setEnv(&env, "RT_OS", config.SystemInfo.OS)
	setEnv(&env, "RT_OS_VER", config.SystemInfo.Version)
	setEnv(&env, "RT_LABELS", labels)
	setEnv(&env, "RT_TEST_NAME", name)
	setEnv(&env, "RT_LIB", libDir)
	setEnv(&env, "RT_RESULTS", config.LogDir)
	if executable == shExecutable {
		envPath := os.Getenv("PATH")
		setEnv(&env, "PATH", fmt.Sprintf("%s:%s", utilsDir, envPath))
		if runtime.GOOS == "windows" {
			setEnv(&env, "MSYS_NO_PATHCONV", "1")
		}
	} else {
		envPath := os.Getenv("Path")
		setEnv(&env, "Path", fmt.Sprintf("%s;%s", utilsDir, envPath))
	}

	cmd.Env = env
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

// setEnv sets or appends key=value in the environment variable passed in
func setEnv(env *[]string, key, value string) {
	for i, e := range *env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			// ignore potentially malformed environment variables
			continue
		}
		if parts[0] == key {
			(*env)[i] = fmt.Sprintf("%s=%s", key, value)
			return
		}
	}

	*env = append(*env, fmt.Sprintf("%s=%s", key, value))
}
