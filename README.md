# Nagios-to-Prometheus Script Executor

## Description

The goal of this tool is to provide a certain level of compatiblity when migrating your monitoring system from Nagios to Prometheus. This works by executing all scripts within the specified directory, and then capturing either the exit code or the standard out (if only a number) to then present it with a metric name in a Prometheus format.  An additional metric **node_n2p_script_executor_lastrun** is also written to a file called `node_n2p_script_executor.prom`.  This is used to determine the time of the last execution of the executor.  The goal is to help identify if generated metrics in the files are stale or not.


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
  -c, --config string         The Path to the config file
  -s, --simulate              Simulate and ouput series to stdout only.
```

*See [sample-config.yml](conf/sample-config.yml) for config example.*


**n2p-script-executor version** (only returns version info and author)


## Sample Script Output

```
node_nagios_script_check_dummy 2.000000
node_nagios_script_check_python_test 0.000000
node_nagios_script_check_google_http 0.000000
node_nagios_script_check_syslogd_running 1.000000
node_nagios_script_check_prom_website_up 0.000000
node_nagios_script_check_google_ping 0.000000
node_n2p_script_executor_lastrun 1589537621319
```

## Building From Source

TODO

## Author

Alain Lefebvre <hartfordfive'at'gmail.com>
