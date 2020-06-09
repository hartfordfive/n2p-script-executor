package executor

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/hartfordfive/n2p-script-executor/lib"

	log "github.com/sirupsen/logrus"
)

var (
	// Debug is used to enable debug logging
	Debug = false
)

// ExecutionResult is a struct generated with the execution result of the script
type ExecutionResult struct {
	ScriptPath string
	ScriptName string
	Metrics    []Metric
	Error      error
}

// Metric is the end result of an execution which contains a metric name, labels and its value
type Metric struct {
	Name   string
	Labels map[string]string
	Value  float64
}

// Script is the struct describing the script to be executed
type Script struct {
	Name               string            `yaml:"name"`
	OutputType         string            `yaml:"output_type"`
	Path               string            `yaml:"path"`
	OverrideMetricName string            `yaml:"override_metric_name"`
	Labels             map[string]string `yaml:"labels"`
	MetricsRegex       string            `yaml:"metrics_regex"`
}

// Config is the struct that maps to the yaml configuration
type Config struct {
	Scripts []Script `yaml:"scripts"`
}

// LoadConfig loads the yaml config from the specified file path
func LoadConfig(path string) (*Config, error) {
	var conf Config
	content, err := ioutil.ReadFile(path)
	err = yaml.Unmarshal(content, &conf)
	//log.Info("Load():", conf)
	if err != nil {
		return &conf, err
	}
	return &conf, err
}

// Print is used to print the config to stdout
func (c *Config) Print() {
	log.Println(c)
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
func RunScript(script Script, timeout int) ExecutionResult {

	if len(script.Labels) == 0 {
		script.Labels = map[string]string{}
	}
	script.Labels["script"] = script.Path

	if script.OutputType == "checkpoint" {
		return ExecutionResult{
			ScriptPath: script.Path,
			ScriptName: script.Name,
			Metrics: []Metric{
				Metric{
					Name:   lib.GetScriptName(script.Path),
					Labels: script.Labels,
					Value:  float64(time.Now().UnixNano() / int64(time.Millisecond)),
				},
			},
		}
	} else if script.OutputType == "exit_code" {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		log.Trace("Running: ", script.Path)
		cmd := exec.CommandContext(ctx, "/bin/bash", "-c", script.Path)
		var waitStatus syscall.WaitStatus
		execErr := cmd.Run()

		if ctx.Err() == context.DeadlineExceeded {
			return ExecutionResult{
				Error: errors.New("Script execution deadline exceeded"),
			}
		}
		if execErr != nil {
			re := regexp.MustCompile("exit status ([0-9])+")
			match := re.FindStringSubmatch(execErr.Error())
			if len(match) >= 1 {
				i, _ := strconv.Atoi(match[1])
				return ExecutionResult{
					ScriptPath: script.Path,
					ScriptName: script.Name,
					Metrics: []Metric{
						Metric{
							Name:   lib.GetScriptName(script.Path),
							Labels: script.Labels,
							Value:  float64(i),
						},
					},
				}

			}
			log.Error(execErr.Error())
			return ExecutionResult{
				Error: errors.New("Could not get exit code"),
			}
		}
		if exitError, ok := execErr.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			return ExecutionResult{
				ScriptPath: script.Path,
				ScriptName: script.Name,
				Metrics: []Metric{
					Metric{
						Name:   lib.GetScriptName(script.Path),
						Labels: script.Labels,
						Value:  float64(waitStatus.ExitStatus()),
					},
				},
			}
		}
		// Success
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		return ExecutionResult{
			ScriptPath: script.Path,
			ScriptName: script.Name,
			Metrics: []Metric{
				Metric{
					Name:   lib.GetScriptName(script.Path),
					Labels: script.Labels,
					Value:  float64(waitStatus.ExitStatus()),
				},
			},
		}
	}

	// In this case, it's either a single metrics (stdout) or a multiple string (multi_metric) output, which will use a supplied regex
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	log.Trace("Running: ", script.Path)
	output, err := exec.CommandContext(ctx, "/bin/bash", "-c", script.Path).Output()

	res := strings.TrimSuffix(string(output), "\n")
	log.Debug("Script output: ", res)

	if err != nil {
		return ExecutionResult{
			Error: fmt.Errorf("Could not get output: %v", err),
		}
	}

	if script.OutputType == "stdout" {
		f, err := strconv.ParseFloat(res, 64)
		if err != nil {
			return ExecutionResult{
				Error: errors.New("Could not parse script output as float"),
			}
		}
		return ExecutionResult{
			ScriptPath: script.Path,
			ScriptName: script.Name,
			Metrics: []Metric{
				Metric{
					Name:   lib.GetScriptName(script.Path),
					Labels: script.Labels,
					Value:  f,
				},
			},
		}
	}

	// In this case, it's an output with potentially multiple metrics,
	// each parsed by a regex
	log.Debugf("Capturing metrics with regex: %s", script.MetricsRegex)
	captures, err := lib.ReturnRegexCaptures(script.MetricsRegex, res)
	if err != nil {
		return ExecutionResult{
			Error: errors.New("Could not parse output with regex"),
		}
	}

	metrics := make([]Metric, len(captures))
	i := 0
	for k, v := range captures {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			continue
		}
		metrics[i] = Metric{
			Name:   fmt.Sprintf("%s_%s", lib.GetScriptName(script.Path), k),
			Labels: script.Labels,
			Value:  f,
		}
		i++
	}
	return ExecutionResult{
		ScriptPath: script.Path,
		ScriptName: script.Name,
		Metrics:    metrics,
	}
}
