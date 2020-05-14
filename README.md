# Nagios-to-Prometheus Script Executor

## Description

The goal of this tool is to provide a certain level of compatiblity when migrating your monitoring system from Nagios to Prometheus. 


## Command Usage

**n2p-script-executor**
```
Available Commands:
  help        Help about any command
  run         Run the script execution
  version     Show version

Flags:
  -h, --help   help for n2p-script-executor
```

**n2p-script-executor run**
```
Flags:
  -h, --help                  help for run
  -l, --log-level string      Enable debug logging.
  -o, --output-file string    Path to the file which the data will be written to, which will in turn be read by the textfile collector module.
  -p, --scripts-path string   Path where the active scripts are located
  -s, --simulate              Simulate only, don't write metrics to output textfile.
```

**n2p-script-executor version** (only returns version info and author)


## Building From Source

TODO

## Author

Alain Lefebvre <hartfordfive@gmail.com>
