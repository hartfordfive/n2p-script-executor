package lib

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/renameio"
	"github.com/hartfordfive/n2p-script-executor/version"

	//"github.com/hartfordfive/n2p-script-executor/executor"
	log "github.com/sirupsen/logrus"
)

// Metric is the end result of an execution which contains a metric name, labels and its value
type Metric struct {
	Name   string
	Labels map[string]string
	Value  float64
	Type   string
	Help   string
}

var seriesPrefix = "n2p_script_exec"

// SetSeriesPrefix updates the prefix of the created series.
func SetSeriesPrefix(prefix string) {
	if len(prefix) >= 1 {
		seriesPrefix = prefix
	}
}

func (m Metric) String(addHelp bool) string {
	output := ""
	if addHelp {
		output += fmt.Sprintf("# TYPE %s_%s %s\n", seriesPrefix, m.Name, m.Type)
		output += fmt.Sprintf("# HELP %s_%s %s\n", seriesPrefix, m.Name, m.Help)
	}
	if len(m.Labels) >= 1 {
		flattendLabels := []string{}
		for k, v := range m.Labels {
			flattendLabels = append(flattendLabels, fmt.Sprintf("%s=\"%s\"", k, v))
		}
		if ValueCanBeInt(m.Value) {
			output += fmt.Sprintf("%s_%s{%s} %s", seriesPrefix, m.Name, strings.Join(flattendLabels, ", "), convertToIntString(m.Value))
		} else {
			output += fmt.Sprintf("%s_%s{%s} %f", seriesPrefix, m.Name, strings.Join(flattendLabels, ", "), m.Value)
		}

	} else {

		if ValueCanBeInt(m.Value) {
			output += fmt.Sprintf("%s_%s %s", seriesPrefix, m.Name, convertToIntString(m.Value))
		} else {
			output += fmt.Sprintf("%s_%s %f", seriesPrefix, m.Name, m.Value)
		}
	}
	//output += "\n"
	return output
}

// IsValidMetricName returns if a metric name is valid or not
func (m Metric) IsValidMetricName() bool {
	if match, _ := regexp.MatchString("[a-zA-Z_:][a-zA-Z0-9_:]*", m.Name); match {
		return true
	}
	return false
}

// ValidSeriesLabels returns if a metric name is valid or not
func (m Metric) ValidSeriesLabels() bool {
	for lblName := range m.Labels {
		if match, _ := regexp.MatchString("[a-zA-Z_][a-zA-Z0-9_]*", lblName); !match {
			return false
		}
	}
	return true
}

// GetScriptName returns the script name without the extension
func GetScriptName(path string) string {
	reg, err := regexp.Compile("[^A-Za-z0-9_]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(strings.Split(filepath.Base(path), ".")[0], "_")
}

// WriteToFile dumps the data to the destination file atomically
func WriteToFile(file string, data string) bool {
	write := func(data string) error {
		t, err := renameio.TempFile("", file)
		if err != nil {
			return err
		}
		defer t.Cleanup()
		if _, err := fmt.Fprintf(t, data); err != nil {
			return err
		}
		return t.CloseAtomicallyReplace()
	}
	if err := write(data); err != nil {
		return false
	}
	return true
}

// ReturnRegexCaptures accepts a regex pattern and returns a map with the matches
func ReturnRegexCaptures(re, str string) (map[string]string, error) {
	r := regexp.MustCompile(re)
	matches := r.FindStringSubmatch(str)
	groups := r.SubexpNames()
	if len(r.FindStringSubmatch(str)) == 0 {
		return nil, errors.New("No matches")
	}
	res := map[string]string{}
	for i, m := range matches {
		if i == 0 {
			continue
		}
		res[groups[i]] = m
	}
	return res, nil
}

// StringIsInSlice returns true if the string is found in the slice.
// Acceptable for search small slices only as it's time comoplexity is O(n)
func StringIsInSlice(item string, list []string) bool {
	for _, elem := range list {
		if item == elem {
			return true
		}
	}
	return false
}

// ValueCanBeInt is used to determine if the float value can really be represented as an int
func ValueCanBeInt(val float64) bool {
	if val-math.Trunc(val) == 0 {
		return true
	}
	return false
}

func convertToIntString(val float64) string {
	return strconv.FormatInt(int64(val), 10)
}

// GenerateSeries takes the array of TextfileCollectorMetric and writes them to the destination file
func GenerateSeries(metrics []Metric, execSuccess []string) string {

	data := ""
	typeHelpLine := map[string]bool{}
	addHelp := false

	for _, metric := range metrics {
		if _, ok := typeHelpLine[metric.Name]; !ok {
			addHelp = true
			typeHelpLine[metric.Name] = true
		}
		log.Debugf("Adding metric: %s", metric.String(addHelp))
		if !metric.IsValidMetricName() {
			log.Warnf("Metric %s has an invalid name. Skipping it.", metric.Name)
		} else if !metric.ValidSeriesLabels() {
			log.Warnf("Metric %s has invalid labels. Removing labels.")
			metric.Labels = map[string]string{}
		} else {
			log.Debugf("Metric name '%s' is valid!", metric.Name)
			data += metric.String(addHelp) + "\n"
			addHelp = false
		}
	}

	data += fmt.Sprintf("# TYPE %s_lastrun counter\n", seriesPrefix)
	data += fmt.Sprintf("# HELP %s_lastrun Time when the script was last executed\n", seriesPrefix)
	for _, script := range execSuccess {
		data += generateScriptCheckpointMetric(script) + "\n"
	}
	data += GenerateExecutorCheckpointMetric() + "\n"

	data += fmt.Sprintf("# TYPE %s_build_info counter\n", seriesPrefix)
	data += fmt.Sprintf("# HELP %s_build_info Build information of the script executor\n", seriesPrefix)
	data += fmt.Sprintf(
		"%s_build_info{version=\"%s\", commit_hash=\"%s\", build_date=\"%s\", go_version=\"%s\"} 1\n",
		seriesPrefix,
		version.Version,
		version.CommitHash,
		version.BuildDate,
		runtime.Version(),
	)

	return data
}

func generateScriptCheckpointMetric(scriptName string) string {
	return fmt.Sprintf(
		"%s_lastrun{script=\"%s\"} %d",
		seriesPrefix,
		scriptName,
		(time.Now().UnixNano() / int64(time.Millisecond)),
	)
}

// GenerateExecutorCheckpointMetric is ...
func GenerateExecutorCheckpointMetric() string {
	return fmt.Sprintf(
		"%s_last_execution %d",
		seriesPrefix,
		(time.Now().UnixNano() / int64(time.Millisecond)),
	)
}
