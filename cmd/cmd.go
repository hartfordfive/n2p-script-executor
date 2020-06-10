package cmd

import (
	"fmt"
	"os"

	"github.com/hartfordfive/n2p-script-executor/executor"
	"github.com/hartfordfive/n2p-script-executor/lib"
	"github.com/hartfordfive/n2p-script-executor/logging"
	"github.com/hartfordfive/n2p-script-executor/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Execute executes the root command.
func Execute() error {
	return entry.Execute()
}

var (
	FlagOutputFile string
	FlagConfig     string
	FlagLogLevel   string
	FlagSimulate   bool
)

var (
	entry = &cobra.Command{
		Use:   "n2p-script-executor",
		Short: "Application to execute legacy Nagios scripts",
		Long: `Nagios-to-Prometheus Script Executor is an application
that allows you to execute legacy Nagios scripts and capture the exit
code to then write it, in Prometheus format, to a file to then be
picked up by the textfile collector module of the node_exporter.`,
	}
)

func init() {
	RunCmd.Flags().StringVarP(&FlagOutputFile, "output-file", "o", "", "Path to the file which the data will be written to, which will in turn be read by the textfile collector module.")
	RunCmd.Flags().StringVarP(&FlagConfig, "config", "c", "", "Path to the config")
	RunCmd.Flags().StringVarP(&FlagLogLevel, "log-level", "l", "", "Enable debug logging.")
	RunCmd.Flags().BoolVarP(&FlagSimulate, "simulate", "s", false, "Simulate only, don't write metrics to output textfile.")
	entry.AddCommand(RunCmd, VersionCmd)
}

// RunCmd is used to initialize the "run" sub-command under the n2p-script-executor
var RunCmd = &cobra.Command{
	Use:   "run ",
	Short: "Run the script execution",
	Long:  `Runs the script execution, which will run all scripts in the specified directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		logging.SetLogLevel(FlagLogLevel)
		run()
		os.Exit(0)

	},
}

// VersionCmd is used to initialize the "version" sub-command under the n2p-script-executor
var VersionCmd = &cobra.Command{
	Use:   "version ",
	Short: "Show version",
	Long:  `Show the version and exit.`,
	Run: func(cmd *cobra.Command, args []string) {
		version.PrintVersion()
		os.Exit(0)
	},
}

func run() {

	cnf, err := executor.LoadConfig(FlagConfig)
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	numWorkers := 4
	if len(cnf.Scripts) < 4 {
		numWorkers = len(cnf.Scripts)
	}

	work := executor.NewWorkQueue(4, numWorkers, len(cnf.Scripts))
	log.Info("Starting script execution workers...")
	work.Process()

	log.Info("Submitting scripts to be executed")
	validOutputTypes := []string{"exit_code", "stdout", "multi_metric", "raw_series"}
	for _, s := range cnf.Scripts {
		if !lib.StringIsInSlice(s.OutputType, validOutputTypes) {
			log.Error("Invalid script output type: ", s.OutputType)
			continue
		}
		work.SubmitTask(s)
	}

	var series []lib.Metric
	var execSuccess = make([]string, 0, len(cnf.Scripts))

	go func(execSuccess *[]string) {
		log.Info("Waiting for results...")
		for res := range work.ResultsChan {
			for _, metric := range res.Metrics {
				series = append(series, metric)
			}
			*execSuccess = append(*execSuccess, res.ScriptPath)
			log.Debug("Decrementing waitgroup")
			work.Wg.Done()
		}
		log.Info("Done processing results")

	}(&execSuccess)

	log.Info("Waiting for all script executions to be completed...")
	work.Wg.Wait()

	seriesOutput := lib.GenerateSeries(series, execSuccess)

	log.Infof("Writing resulting series to %s", FlagOutputFile)
	if !FlagSimulate {
		if len(series) >= 1 {
			// Write the series to the output file
			lib.WriteToFile(FlagOutputFile, seriesOutput)
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
