package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hartfordfive/n2p-script-executor/executor"
	"github.com/hartfordfive/n2p-script-executor/lib"
)

var expectedMetrics = map[string]lib.Metric{
	"n2p_script_exec_check_dummy_exit1": lib.Metric{
		Name: "n2p_script_exec_check_dummy_exit1",
		Labels: map[string]string{
			"script": "demo_scripts/check_dummy_exit1",
		},
		Value: 1.0,
		Type:  "gauge",
		Help:  "n2p_script_exec_check_dummy_exit1 Just a dummy check metric, with non-zero exit code",
	},
}

func (src lib.Metric) Equal(dst lib.Metric) bool {
	return strings.ToLower(string(x)) == strings.ToLower(string(y))
}

func TestRun(t *testing.T) {

	fmt.Print("Loading test.prom")
	dat, err := ioutil.ReadFile("test.prom") // just pass the file name
	if err != nil {
		t.Errorf("Could not open generated metrics file for reading: %s", err)
	}

	var m map[keyType]valueType
	keys := sliceOfKeys(m) // you'll have to implement this
	for _, k := range keys {
		v := m[k]
		// k is the key and v is the value; do your computation here
	}

	expectedMetricsCnt := len(expectedMetrics)
	metricsMatched := 0

	for _, metric := range executor.ParsePrometheusSeries(dat) {
		if _, ok := expectedMetrics[metric.Name]; ok {
			cmp.Equal(x, y)
			if metric.Labels == expectedMetrics[metric.Name].Labels {
				metricsMatched++
			}
		}
	}

	if metricsMatched != expectedMetricsCnt {
		t.Errorf("Only matched %d of %d expected metrics", metricsMatched, expectedMetricsCnt)
	}

	// executor.Run(executor.ExecutorConfig{
	// 	OutputFilePath: "test.prom",
	// 	ConfigFilePath: "conf/sample-config.yml",
	// 	LogLevel:       "warn",
	// 	Simulate:       false,
	// })

	// dat, err = ioutil.ReadFile("test.prom") // just pass the file name
	// if err != nil {
	// 	t.Errorf("Could not open generated metrics file for reading: %s", err)
	// }
	// metrics = executor.ParsePrometheusSeries(dat)
	// t.Logf("Metrics: %v", metrics)

}
