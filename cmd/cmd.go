package cmd

import (
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
	FlagScriptsDir string
	FlagOutputFile string
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
	RunCmd.Flags().StringVarP(&FlagScriptsDir, "scripts-path", "p", "", "Path where the active scripts are located")
	RunCmd.Flags().StringVarP(&FlagOutputFile, "output-file", "o", "", "Path to the file which the data will be written to, which will in turn be read by the textfile collector module.")
	RunCmd.Flags().StringVarP(&FlagLogLevel, "log-level", "l", "", "Enable debug logging.")
	RunCmd.Flags().BoolVarP(&FlagSimulate, "simulate", "s", false, "Simulate only, don't write metrics to output textfile.")
	entry.AddCommand(RunCmd, VersionCmd)
}

// RunCmd is used to initialize the "run" sub-command under the n2p-script-executor
var RunCmd = &cobra.Command{
	Use:   "run ",
	Short: "Run the script execution",
	Long:  `Runs the script execution, which will run all scripts in the specified directory.`,
	//Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logging.SetLogLevel(FlagLogLevel)
		log.Debug("Running the script executor.")
		scripts, err := executor.GetScripts(FlagScriptsDir)
		if err != nil {
			log.Error(err)
		}

		numWorkers := 4
		if len(scripts) < 4 {
			numWorkers = len(scripts)
		}

		work := executor.NewWorkQueue(4, numWorkers, len(scripts))
		log.Info("Starting script execution workers...")
		work.Process()

		log.Info("Submitting scripts to be executed")
		for _, s := range scripts {
			work.SubmitTask(s)
		}

		var series []lib.TextfileCollectorMetric
		go func() {
			log.Info("Waiting for results...")

			for elem := range work.ResultsChan {
				series = append(series, lib.TextfileCollectorMetric{
					OriginScript: lib.GetScriptName(elem.ScriptPath),
					Labels:       nil,
					Value:        elem.Metric,
				})
				log.Debugf("Script: %v, Exit code: %d, Error: %v", elem.ScriptPath, elem.Metric, elem.Error)
				work.Wg.Done()
			}
			log.Info("Done processing results")

		}()

		work.Wg.Wait()
		if !FlagSimulate {
			if len(series) >= 1 {
				// Write the series to the output file
				lib.WriteSeriesToFile(series, FlagOutputFile)
				// Write a custom metric that indicates the unix millisecond timestamp at which the last time the latest metrics were written
				lib.WriteCheckpointMetric(FlagOutputFile)
				os.Exit(0)

			}
			os.Exit(1)
		}

		lib.WriteSeriesToStdOut(series)
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
