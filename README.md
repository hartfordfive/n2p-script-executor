# Nagios-to-Prometheus Script Executor

## Description

The goal of this tool is to provide simple method of obtaining metrics from the execution of multiple complex commands as well as provide a certain level of compatiblity when migrating from a legacy script-based monitoring system such as Nagios to Prometheus. 


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


## Configuration

`series_prefix` : The prefix to be used for all resulting series (default = n2p_script_exec) 
`script.[X].name` : The identifier name to use for this script.  By default, this is added to the metric name. (mandatory)
`script.[X].type` : The type of prometheus metric. See Prometheus types for all types. (mandatory)
`script.[X].help` : The help message to add as the # HELP string (optional)
`script.[X].path` : The path of the script (mandatory)
`script.[X].output_type` : One of exit_code, raw_series, multi_metric, stdout
`script.[X].override_metric_name` : If specified and valid, will be used as the series suffix (optional)
`script.[X].labels.<LABEL_NAME>` : Additional list of label/value pairs to be included to to each resulting series produced by the script (optional)


## Script Output Types

The following four output types are available:

**stdout** : Stdout will be used as value
**exit_code** : Exit code will be used as value
**multi_metric** : 
**raw_series** : Prometheus formated (same as OpenTSDB) metrics will be included as is



## Examples

Please execute the scripts in the examples directory for sample metrics output.

to write series to a file:
```bash
<BINARY_PATH> run -c examples/sample-config.yml -l info -o test.out
```

or only write series to stdout:
```bash
<BINARY_PATH> run -c examples/sample-config.yml -l info -s
```


## Common Metrics

The following metrics are added on each execution:

`script_exec_build_info` : The build info of the executor, including the author, version, build time and commit hash.
`script_exec_lastrun`  : The last time a given configured script was executed.  Please note this metric only appears if a script was executed.
`script_exec_script_last_run_success`  : The last time, as a millisecond timestamp, a configured script was successfully executed.
`script_exec_last_execution`  :  The last time, as a millisecond timestamp, a configured script was executed.  This includes both successful and uncessful executions.
`script_exec_script_last_execution_time_ms` : The total of the last execution time of a configured script in milliseconds.
`script_exec_script_loaded` : Indicates if a script has been properly detected and scheduled for execution.


## Sample Executor Output

```
# TYPE script_exec_check_dummy gauge
# HELP script_exec_check_dummy Just a dummy check metric, with zero exit code
script_exec_check_dummy{script="examples/check_dummy"} 0
# TYPE script_exec_my_custom_app2_total_requests
# HELP script_exec_my_custom_app2_total_requests
script_exec_my_custom_app2_total_requests{shard="users01", instance="web01"} 46223
script_exec_my_custom_app2_total_requests{instance="app01", shard="transactions01"} 2466
script_exec_my_custom_app2_total_requests{instance="web03", shard="users02"} 48665
# TYPE script_exec_http_site gauge
# HELP script_exec_http_site Check to verify if google.com is up
script_exec_http_site{site_category="search_engine", protocol="https", url="www.google.com", script="examples/check_google_http"} 0
# TYPE script_exec_script_loaded gauge
# HELP script_exec_script_loaded indicates that a script has been identified to be executed
script_exec_script_loaded{script="examples/check_dummy"} 1
# TYPE script_exec_script_last_execution_time_ms gauge
# HELP script_exec_script_last_execution_time_ms indicates the number of milliseconds it has taken to execute the script
script_exec_script_last_execution_time_ms{script="examples/check_dummy"} 14
script_exec_script_loaded{script="examples/check_raw_metrics_exit1"} 1
script_exec_script_last_execution_time_ms{script="examples/check_raw_metrics_exit1"} 15
script_exec_script_loaded{script="examples/check_google_http"} 1
script_exec_script_last_execution_time_ms{script="examples/check_google_http"} 41
# TYPE script_exec_script_last_run_success gauge
# HELP script_exec_script_last_run_success iindicates when a script was last executed successfully
script_exec_script_last_run_success{script="examples/check_dummy"} 1
script_exec_script_last_run_success{script="examples/check_raw_metrics_exit1"} 1
script_exec_script_last_run_success{script="examples/check_google_http"} 1
# TYPE script_exec_lastrun counter
# HELP script_exec_lastrun Time when the script was last executed
script_exec_lastrun{script="examples/check_dummy"} 1614101626811
script_exec_lastrun{script="examples/check_raw_metrics_exit1"} 1614101626811
script_exec_lastrun{script="examples/check_google_http"} 1614101626811
script_exec_last_execution 1614101626811
# TYPE script_exec_build_info counter
# HELP script_exec_build_info Build information of the script executor
script_exec_build_info{version="0.6.1", commit_hash="07d793359fd826383be81d721bb1da91e23d351c", build_date="2021-02-23", go_version="go1.15.5"} 1
```

## Building

To compile the binaries for both Linux and OSX
```
make build-all
```

To compile the binaries and additionally create the tar.gz archives
```
make build-release
```

*See makefile for all options*


## Author

Alain Lefebvre <hartfordfive'at'gmail.com>


