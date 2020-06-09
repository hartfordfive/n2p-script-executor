package lib

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const tagName = "textfile_collector"

// TextfileCollectorMetric is a struct which will be serialized to be written to a file
// to be picked up by the node_exporter
type TextfileCollectorMetric struct {
	Name   string
	Labels map[string]string
	Value  float64
}

func (tc TextfileCollectorMetric) String() string {
	if len(tc.Labels) >= 1 {
		flattendLabels := []string{}
		for k, v := range tc.Labels {
			flattendLabels = append(flattendLabels, fmt.Sprintf("%s=\"%s\"", k, v))
		}

		if ValueCanBeInt(tc.Value) {
			return fmt.Sprintf("n2p_script_exec_%s{%s} %s", tc.Name, strings.Join(flattendLabels, ", "), convertToIntString(tc.Value))
		}
		return fmt.Sprintf("n2p_script_exec_%s{%s} %f", tc.Name, strings.Join(flattendLabels, ", "), tc.Value)
	}

	if ValueCanBeInt(tc.Value) {
		return fmt.Sprintf("n2p_script_exec_%s %s", tc.Name, convertToIntString(tc.Value))
	}
	return fmt.Sprintf("n2p_script_exec_%s %f", tc.Name, tc.Value)
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
func GenerateSeries(metrics []TextfileCollectorMetric, execSuccess []string) string {
	data := ""
	for _, metric := range metrics {
		log.Debugf("Adding metric: %s", metric.String())
		data += metric.String() + "\n"
	}
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

func GenerateExecutorCheckpointMetric() string {
	return fmt.Sprintf(
		"n2p_script_exec_lastrun %d",
		(time.Now().UnixNano() / int64(time.Millisecond)),
	)
}
