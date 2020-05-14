package main

import (
	"os"

	"github.com/hartfordfive/n2p-script-executor/cmd"
	"github.com/hartfordfive/n2p-script-executor/logging"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&logging.LogFormatPlain)
	log.SetOutput(os.Stdout)
	//log.SetReportCaller(true)
}

func main() {

	// entry := &cobra.Command{Use: "n2p-script-executor"}
	// cmd.RunCmd.Flags().StringVarP(&cmd.FlagScriptsDir, "scripts-path", "p", "", "Path where the active scripts are located")
	// cmd.RunCmd.Flags().StringVarP(&cmd.FlagOutputFile, "output-file", "o", "", "Path to the file which the data will be written to, which will in turn be read by the textfile collector module.")
	// cmd.RunCmd.Flags().StringVarP(&cmd.FlagLogLevel, "log-level", "l", "", "Enable debug logging.")
	// cmd.RunCmd.Flags().BoolVarP(&cmd.FlagDryRun, "dry-run", "d", false, "Dry-run only, don't write metrics to textfile.")
	// entry.AddCommand(cmd.RunCmd, cmd.VersionCmd)

	cmd.Execute()
}
