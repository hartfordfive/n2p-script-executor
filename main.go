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

	cmd.Execute()
}
