package lib

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/renameio"
	log "github.com/sirupsen/logrus"
)

// GetScriptName returns the script name without the extension
func GetScriptName(path string) string {
	reg, err := regexp.Compile("[^A-Za-z0-9_]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(strings.Split(filepath.Base(path), ".")[0], "_")
}

// WriteToFile dumps the data to the destination file atomically
func WriteToFile(file string, data string) bool {
	write := func(data string) error {
		t, err := renameio.TempFile("", file)
		if err != nil {
			return err
		}
		defer t.Cleanup()
		if _, err := fmt.Fprintf(t, data); err != nil {
			return err
		}
		return t.CloseAtomicallyReplace()
	}
	if err := write(data); err != nil {
		return false
	}
	return true
}
