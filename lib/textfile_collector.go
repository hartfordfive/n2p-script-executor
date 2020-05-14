package lib

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const tagName = "textfile_collector"

// TextfileCollectorMetric is a struct which will be serialized to be written to a file
// to be picked up by the node_exporter
type TextfileCollectorMetric struct {
	OriginScript string
	Labels       []string
	Value        float64
}

func (tc TextfileCollectorMetric) String() string {
	if len(tc.Labels) >= 1 {
		return fmt.Sprintf("node_nagios_script_%s{%s} %f", GetScriptName(tc.OriginScript), strings.Join(tc.Labels, ","), tc.Value)
	}
	return fmt.Sprintf("node_nagios_script_%s %f", GetScriptName(tc.OriginScript), tc.Value)
}

// WriteSeriesToFile takes the array of TextfileCollectorMetric and writes them to the destination file
func WriteSeriesToFile(metrics []TextfileCollectorMetric, file string) {
	log.Debugf("Writing series to file: %s", file)
	data := ""
	for _, metric := range metrics {
		log.Debug("Adding metric: %s", metric.String())
		data += metric.String()
		data += "\n"
	}
	WriteToFile(file, data)
}

func WriteCheckpointMetric(file string) {
	WriteToFile(filepath.Join(filepath.Dir(file), "node_n2p_script_executor.txt"), fmt.Sprintf("node_n2p_script_executor_lastrun %d", (time.Now().UnixNano()/int64(time.Millisecond))))
}
