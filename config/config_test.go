package config

import (
	"testing"
)

var expectedConf = &Config{
	SeriesPrefix: "script_executor",
	Scripts: []Script{
		Script{
			"check_dummy",
			"gauge",
			"Just a dummy check metric, with zero exit code",
			"examples/check_dummy",
			"exit_code",
		},
	},
}

func TestConfig(t *testing.T) {

	conf, err := LoadConfig("examples/sample-config.yml")
	if err != nil {
		t.Errorf("Could not open config file: %s", err)
	}

}
