package lib

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/renameio"

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

func (m Metric) String() string {
	output := ""
	output += fmt.Sprintf("# TYPE n2p_script_exec_%s %s\n", m.Name, m.Type)
	output += fmt.Sprintf("# HELP n2p_script_exec_%s %s\n", m.Name, m.Help)
	if len(m.Labels) >= 1 {
		flattendLabels := []string{}
		for k, v := range m.Labels {
			flattendLabels = append(flattendLabels, fmt.Sprintf("%s=\"%s\"", k, v))
		}
		if ValueCanBeInt(m.Value) {
			output += fmt.Sprintf("n2p_script_exec_%s{%s} %s", m.Name, strings.Join(flattendLabels, ", "), convertToIntString(m.Value))
		} else {
			output += fmt.Sprintf("n2p_script_exec_%s{%s} %f", m.Name, strings.Join(flattendLabels, ", "), m.Value)
		}

	} else {

		if ValueCanBeInt(m.Value) {
			output += fmt.Sprintf("n2p_script_exec_%s %s", m.Name, convertToIntString(m.Value))
		} else {
			output += fmt.Sprintf("n2p_script_exec_%s %f", m.Name, m.Value)
		}
	}
	//output += "\n"
	return output
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
	for _, metric := range metrics {
		log.Debugf("Adding metric: %s", metric.String())
		data += metric.String() + "\n"
	}
	data += "# TYPE n2p_script_exec_lastrun counter\n"
	data += "# HELP n2p_script_exec_lastrun Time when the script was last executed\n"
	for _, script := range execSuccess {
		data += generateScriptCheckpointMetric(script) + "\n"
	}
	data += GenerateExecutorCheckpointMetric() + "\n"

	return data
}

func generateScriptCheckpointMetric(scriptName string) string {
	return fmt.Sprintf(
		"n2p_script_exec_lastrun{script=\"%s\"} %d",
		scriptName,
		(time.Now().UnixNano() / int64(time.Millisecond)),
	)
}

// GenerateExecutorCheckpointMetric is ...
func GenerateExecutorCheckpointMetric() string {
	return fmt.Sprintf(
		"n2p_script_exec_lastrun %d",
		(time.Now().UnixNano() / int64(time.Millisecond)),
	)
}
