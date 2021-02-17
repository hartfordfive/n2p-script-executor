package executor

import (
	"fmt"
	"os"

	"github.com/hartfordfive/n2p-script-executor/config"
	"github.com/hartfordfive/n2p-script-executor/lib"
	log "github.com/sirupsen/logrus"
)

// ExecutorConfig config specifies the config for the executore
type ExecutorConfig struct {
	OutputFilePath string
	ConfigFilePath string
	LogLevel       string
	Simulate       bool
}

// Run runs the executor
func Run(cfg ExecutorConfig) {

	cnf, err := config.Load(cfg.ConfigFilePath)
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	numWorkers := 4
	if len(cnf.Scripts) < 4 {
		numWorkers = len(cnf.Scripts)
	}

	if len(cnf.SeriesPrefix) >= 1 {
		lib.SetSeriesPrefix(cnf.SeriesPrefix)
	}

	work := NewWorkQueue(4, numWorkers, len(cnf.Scripts))
	log.Info("Starting script execution workers...")
	work.Process()

	log.Info("Submitting scripts to be executed")
	for _, s := range cnf.Scripts {
		work.SubmitTask(s)
	}

	var series []lib.Metric
	scriptLoadedSeries := []lib.Metric{}
	scriptExecSuccessSeries := []lib.Metric{}
	var execSuccess = make([]string, 0, len(cnf.Scripts))

	go func(execSuccess *[]string) {
		log.Info("Waiting for results...")

		for res := range work.ResultsChan {

			scriptLoadedSeries = append(scriptLoadedSeries, lib.Metric{
				Name: "script_loaded",
				Labels: map[string]string{
					"script": res.ScriptPath,
				},
				Value: 1.0,
				Type:  "gauge",
				Help:  "indicates that a script has been identified to be executed",
			})

			scriptLoadedSeries = append(scriptLoadedSeries, lib.Metric{
				Name: "script_last_execution_time_ms",
				Labels: map[string]string{
					"script": res.ScriptPath,
				},
				Value: float64(res.TotalExecTime),
				Type:  "gauge",
				Help:  "indicates the number of milliseconds it has taken to execute the script",
			})

			lastRunSuccess := 0.0

			if res.Error == nil {
				for _, metric := range res.Metrics {
					series = append(series, metric)
				}
				*execSuccess = append(*execSuccess, res.ScriptPath)
				lastRunSuccess = 1.0
			}

			scriptExecSuccessSeries = append(scriptExecSuccessSeries, lib.Metric{
				Name: "script_last_run_success",
				Labels: map[string]string{
					"script": res.ScriptPath,
				},
				Value: lastRunSuccess,
				Type:  "gauge",
				Help:  "iindicates when a script was last executed successfully",
			})

			log.Debug("Decrementing waitgroup")
			work.Wg.Done()
		}

		log.Info("Done processing results")

	}(&execSuccess)

	log.Info("Waiting for all script executions to be completed...")
	work.Wg.Wait()

	for _, res := range scriptLoadedSeries {
		series = append(series, res)
	}

	for _, res := range scriptExecSuccessSeries {
		series = append(series, res)
	}

	seriesOutput := lib.GenerateSeries(series, execSuccess)

	if cfg.OutputFilePath != "" {
		log.Infof("Writing resulting series to %s", cfg.OutputFilePath)
	} else {
		log.Info("Writing resulting series to stdout")
	}
	if !cfg.Simulate {
		if len(series) >= 1 {
			// Write the series to the output file
			lib.WriteToFile(cfg.OutputFilePath, seriesOutput)
			os.Exit(0)

		}
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, seriesOutput)
	if len(series) >= 1 {
		os.Exit(0)
	}

	os.Exit(1)
}
