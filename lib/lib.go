package lib

import (
	"errors"
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

// ReturnRegexCaptures accepts a regex pattern and returns a map with the matches
func ReturnRegexCaptures(re, str string) (map[string]string, error) {
	r := regexp.MustCompile(re)
	matches := r.FindStringSubmatch(str)
	groups := r.SubexpNames()
	if len(r.FindStringSubmatch(str)) == 0 {
		return nil, errors.New("No matches")
	}
	res := map[string]string{}
	for i, m := range matches {
		if i == 0 {
			continue
		}
		res[groups[i]] = m
	}
	return res, nil
}

// StringIsInSlice returns true if the string is found in the slice.
// Acceptable for search small slices only as it's time comoplexity is O(n)
func StringIsInSlice(item string, list []string) bool {
	for _, elem := range list {
		if item == elem {
			return true
		}
	}
	return false
}
