package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hartfordfive/n2p-script-executor/config"
	"github.com/hartfordfive/n2p-script-executor/lib"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/textparse"

	log "github.com/sirupsen/logrus"
)

var (
	// Debug is used to enable debug logging
	Debug = false
)

// ExecutionResult is a struct generated with the execution result of the script
type ExecutionResult struct {
	ScriptPath    string
	ScriptName    string
	Metrics       []lib.Metric
	Error         error
	TotalExecTime int64
}

// GetScripts returns the list of scripts in the provided directory when in simple mode
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
func RunScript(script config.Script) ExecutionResult {

	if len(script.Labels) == 0 {
		script.Labels = map[string]string{}
	}
	script.Labels["script"] = script.Path

	timeout, _ := time.ParseDuration(script.Timeout)

	execStart := time.Now()

	if script.OutputType == "exit_code" {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		log.Trace("Running script %s (output type: %s)", script.Path, script.OutputType)
		cmd := exec.CommandContext(ctx, "/bin/bash", "-c", script.Path)
		var waitStatus syscall.WaitStatus
		execErr := cmd.Run()

		duration := time.Since(execStart)
		execTotalMs := duration.Milliseconds()

		if ctx.Err() == context.DeadlineExceeded {
			return ExecutionResult{
				ScriptPath:    script.Path,
				Error:         errors.New("Script execution deadline exceeded"),
				TotalExecTime: execTotalMs,
			}
		}
		if execErr != nil {
			re := regexp.MustCompile("exit status ([0-9])+")
			match := re.FindStringSubmatch(execErr.Error())
			if len(match) >= 1 {
				log.Debugf("Success running script %s (output type: %s)", script.Path, script.OutputType)
				i, _ := strconv.Atoi(match[1])
				return ExecutionResult{
					ScriptPath: script.Path,
					ScriptName: script.Name,
					Metrics: []lib.Metric{
						lib.Metric{
							Name:   lib.GetScriptName(script.Path),
							Labels: script.Labels,
							Value:  float64(i),
							Type:   script.Type,
							Help:   script.Help,
						},
					},
					TotalExecTime: execTotalMs,
				}

			}
			log.Error(execErr.Error())
			return ExecutionResult{
				ScriptPath:    script.Path,
				Error:         errors.New("Could not get exit code"),
				TotalExecTime: execTotalMs,
			}
		}
		if exitError, ok := execErr.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			log.Debugf("Success running script %s (output type: %s)", script.Path, script.OutputType)
			return ExecutionResult{
				ScriptPath: script.Path,
				ScriptName: script.Name,
				Metrics: []lib.Metric{
					lib.Metric{
						Name:   lib.GetScriptName(script.Path),
						Labels: script.Labels,
						Value:  float64(waitStatus.ExitStatus()),
						Type:   script.Type,
						Help:   script.Help,
					},
				},
				TotalExecTime: execTotalMs,
			}
		}
		// Success
		log.Debugf("Success running script %s (output type: %s)", script.Path, script.OutputType)
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		return ExecutionResult{
			ScriptPath: script.Path,
			ScriptName: script.Name,
			Metrics: []lib.Metric{
				lib.Metric{
					Name:   lib.GetScriptName(script.Path),
					Labels: script.Labels,
					Value:  float64(waitStatus.ExitStatus()),
					Type:   script.Type,
					Help:   script.Help,
				},
			},
			TotalExecTime: execTotalMs,
		}
	}

	log.Debugf("Running script %s (output type: %s) with timeout of %v", script.Path, script.OutputType, timeout)

	// The following block of code is used instead of CommandContext as the context
	// doesn't terminate sub-processes.
	// See: https://github.com/golang/go/issues/22485

	cmd := exec.Command("/bin/bash", "-c", script.Path)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	killChanRes := make(chan ExecutionResult, 1)

	ctx := context.Background()
	ctx, cancelTimeout := context.WithCancel(ctx)

	go func(ctx context.Context) {
		select {
		case <-time.After(timeout):
			log.Warnf("timeout of %v has passed for script %s (output type: %s). Killing script",
				timeout,
				script.Path,
				script.OutputType)
			duration := time.Since(execStart)
			execTotalMs := duration.Milliseconds()
			killChanRes <- ExecutionResult{
				ScriptPath:    script.Path,
				Error:         errors.New("Script execution timed out"),
				TotalExecTime: execTotalMs,
			}
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
				log.Warnf("Error killing running script %s (output type: %s): %s",
					script.Path,
					script.OutputType,
					err)
			}
		case <-ctx.Done():
			log.Debugf("timeout canceled for %s (output type: %s)", script.Path, script.OutputType)
		}
	}(ctx)

	output, outErr := cmd.Output()
	if outErr == nil {
		cancelTimeout()
	}

	duration := time.Since(execStart)
	execTotalMs := duration.Milliseconds()

	res := strings.TrimSuffix(string(output), "\n")
	log.Debugf("Script output for %s (output type: %s) : %s",
		script.Path,
		script.OutputType,
		res)

	if outErr != nil && script.OutputType != "raw_series" {
		return ExecutionResult{
			ScriptPath:    script.Path,
			Error:         fmt.Errorf("Could not get output: %v", outErr),
			TotalExecTime: execTotalMs,
		}
	}

	if script.OutputType == "raw_series" {
		metrics := ParsePrometheusSeries(output)
		if len(metrics) == 0 {
			return ExecutionResult{
				ScriptPath: script.Path,
				Error:      fmt.Errorf("No valid series detected in %s", script.Path),
			}
		}
		return ExecutionResult{
			ScriptPath: script.Path,
			ScriptName: script.Name,
			Metrics:    metrics,
		}
	} else if script.OutputType == "stdout" {
		f, err := strconv.ParseFloat(res, 64)
		if err != nil {
			return ExecutionResult{
				ScriptPath:    script.Path,
				Error:         errors.New("Could not parse script output as float"),
				TotalExecTime: execTotalMs,
			}
		}
		return ExecutionResult{
			ScriptPath: script.Path,
			ScriptName: script.Name,
			Metrics: []lib.Metric{
				lib.Metric{
					Name:   lib.GetScriptName(script.Path),
					Labels: script.Labels,
					Value:  f,
					Type:   script.Type,
					Help:   script.Help,
				},
			},
			TotalExecTime: execTotalMs,
		}
	}

	// In this case, it's an output with potentially multiple metrics,
	// each parsed by a regex
	log.Debugf("Capturing metrics from %s (output type: %s) with regex: %s",
		script.Path,
		script.OutputType,
		script.MetricsRegex)
	captures, err := lib.ReturnRegexCaptures(script.MetricsRegex, res)
	if err != nil {
		return ExecutionResult{
			ScriptPath:    script.Path,
			Error:         errors.New("Could not parse output with regex"),
			TotalExecTime: execTotalMs,
		}
	}

	metrics := make([]lib.Metric, len(captures))
	i := 0
	for k, v := range captures {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			continue
		}
		metrics[i] = lib.Metric{
			Name:   fmt.Sprintf("%s_%s", lib.GetScriptName(script.Path), k),
			Labels: script.Labels,
			Value:  f,
			Type:   script.Type,
			Help:   script.Help,
		}
		i++
	}
	return ExecutionResult{
		ScriptPath:    script.Path,
		ScriptName:    script.Name,
		Metrics:       metrics,
		TotalExecTime: execTotalMs,
	}
}

// ParsePrometheusSeries parses metrics provided in rawSeriesOutput.
// In case a bad read occurs (because of an illegal metric format or whatever),
// all metrics read up to this point will be returned. Every metrics present
// after are ignored.
func ParsePrometheusSeries(rawSeriesOutput []byte) []lib.Metric {
	p := textparse.NewPromParser(rawSeriesOutput)

	var resLabels labels.Labels
	labelsMap := map[string]string{}

	metrics := make([]lib.Metric, 0)
	var metricName string

	for {
		entry, err := p.Next()

		// All metrics have been read. Exit.
		if err == io.EOF {
			break
		}

		// True if a bad read occurs. The Parsing gets stuck in this case and no other option than quitting is left.
		if err != nil && entry == textparse.EntryInvalid {
			series, _, _ := p.Series()
			log.Errorf("invalid metric %s : %s\n", string(series), err)
			break
		}

		switch entry {
		case textparse.EntrySeries:
			_, _, v := p.Series()
			p.Metric(&resLabels)
			for _, lbl := range resLabels {
				if strings.Trim(lbl.Value, " ") == "" {
					log.Warnf("Skipping label %s as it has an empty value", lbl.Name)
					continue
				}
				if lbl.Name == "__name__" {
					metricName = lbl.Value
					continue
				}
				labelsMap[lbl.Name] = lbl.Value
			}
			metrics = append(metrics, lib.Metric{
				Name:   metricName,
				Labels: labelsMap,
				Value:  v,
			})
		}
		resLabels = labels.Labels{}
		labelsMap = make(map[string]string)
	}

	return metrics
}
