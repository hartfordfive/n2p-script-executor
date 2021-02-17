package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/hartfordfive/n2p-script-executor/lib"
)

// Script is the struct describing the script to be executed
type Script struct {
	Name               string            `yaml:"name"`
	Timeout            string            `yaml:"timeout"`
	Type               string            `yaml:"type"`
	Help               string            `yaml:"help"`
	OutputType         string            `yaml:"output_type"`
	Path               string            `yaml:"path"`
	OverrideMetricName string            `yaml:"override_metric_name"`
	Labels             map[string]string `yaml:"labels"`
	MetricsRegex       string            `yaml:"metrics_regex"`
}

// Config is the struct that maps to the yaml configuration
type Config struct {
	SeriesPrefix string   `yaml:"series_prefix"`
	Scripts      []Script `yaml:"scripts"`
}

// Load loads the yaml config from the specified file path
func Load(path string) (*Config, error) {
	var conf Config
	content, err := ioutil.ReadFile(path)
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		return &conf, err
	}

	if err := conf.InitAndValidate(); err != nil {
		return nil, err
	}

	return &conf, err
}

// InitAndValidate verifies the config is valid before attempting to run the executor
func (c *Config) InitAndValidate() error {

	validOutputTypes := []string{"exit_code", "stdout", "multi_metric", "raw_series"}

	if len(c.Scripts) == 0 {
		return errors.New("must specify at least one script to execute")
	}

	for i := range c.Scripts {
		_, err := os.Stat(c.Scripts[i].Path)
		if os.IsNotExist(err) {
			return fmt.Errorf("script '%s' does not exist", c.Scripts[i].Path)
		}

		if c.Scripts[i].Timeout == "" {
			tout := &c.Scripts[i].Timeout
			*tout = "10s"
		} else {
			d, err := time.ParseDuration(c.Scripts[i].Timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout duration string '%s' for script '%s'", c.Scripts[i].Timeout, c.Scripts[i].Path)
			}
			if d.Seconds() < 1 {
				return fmt.Errorf("timeout for script '%s' must be >= 1 (value passed: %d)", c.Scripts[i].Path, c.Scripts[i].Timeout)
			}
		}
		if !lib.StringIsInSlice(c.Scripts[i].OutputType, validOutputTypes) {
			return (fmt.Errorf("Invalid script output type: %s", c.Scripts[i].OutputType))
		}
	}

	return nil
}

// Print is used to print the config to stdout
func (c *Config) Print() {
	log.Println(c)
}
