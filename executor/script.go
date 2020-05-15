package executor

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	// Debug is used to enable debug logging
	Debug = false
)

// ExecutionResult is a struct generated with the execution result of the script
type ExecutionResult struct {
	ScriptPath string
	Metric     float64
	Error      error
}

// GetScripts returns the list of scripts in the provided directory
func GetScripts(scriptsDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
		if fi, err := os.Stat(path); err == nil {
			if fi.Mode().IsRegular() {
				log.Debug("Found script:", path)
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		log.Error("Could not get scripts in directory: %v", err.Error())
		return []string{}, err
	}
	log.Debug("Done finding scripts")
	return files, nil
}

// RunScript starts the execution of the script
func RunScript(scriptPath string, metricType string, timeout int) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	log.Trace("Running: ", scriptPath)
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", scriptPath)
	var waitStatus syscall.WaitStatus
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return -1, errors.New("Script execution deadline exceeded")
	}

	if metricType == "stdout" {
		output, err := cmd.Output()
		if err != nil {
			return float64(-1), err
		}
		f, err := strconv.ParseFloat(string(output), 64)
		if err != nil {
			return float64(-1), err
		}
		return f, nil
	}

	if err != nil {
		if err != nil {
			re := regexp.MustCompile("exit status ([0-9])+")
			match := re.FindStringSubmatch(err.Error())
			if len(match) >= 1 {
				i, _ := strconv.Atoi(match[1])
				return float64(i), nil
			}
			log.Error(err.Error())
			return float64(-1), err
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			return float64(waitStatus.ExitStatus()), nil
		}
	}

	// Success
	waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
	return float64(waitStatus.ExitStatus()), nil

}
