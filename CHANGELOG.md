
## 0.2.1
* Added new `raw_series` output type which directly accepts a text representation of prometheus series
* Added new `type` and `help` settings in configuration to give the ability to specify the corresponding `# TYPE` and `# HELP` metric comments

## 0.2.0
* Scripts are now defined via a configuration which also allows you to indicate how to parse a scripts execution result. 
* Added individual checkpoint metric per script so that users have the ability to tell if a given script execution result is stale.
* Renamed output series prefix

## 0.1.0
* Initial Release