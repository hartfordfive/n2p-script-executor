package cmd

import (
	"os"

	"github.com/hartfordfive/n2p-script-executor/executor"
	"github.com/hartfordfive/n2p-script-executor/logging"
	"github.com/hartfordfive/n2p-script-executor/version"
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
		executor.Run(executor.ExecutorConfig{
			OutputFilePath: FlagOutputFile,
			ConfigFilePath: FlagConfig,
			LogLevel:       FlagLogLevel,
			Simulate:       FlagSimulate,
		})
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
