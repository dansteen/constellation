package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/dansteen/constellation/util"
	"github.com/ghodss/yaml"
)

// ProcessFile will process config files for constellation, and return an array of Config objects
// required files are processed as well.  Accepts a filepath and an array of directories to search for
// files that are included via the 'require' stanza (or the file provided)
func ProcessFile(fileName string, includeDirs []string) Config {

	// first make sure the file exists
	filePath, err := findFile(fileName, includeDirs)
	util.Check(err)

	// setup a map to hold our config
	config := Config{}

	// read in the file provided
	data, err := ioutil.ReadFile(filePath)
	util.Check(err)
	err = yaml.Unmarshal(data, &config)
	//util.Check(err)
	fmt.Printf("%+v\n", err)

	// run through and merge any reqired files in
	for _, requirePath := range config.Requires {
		// get containers from the requires and add them to our list
		requireConfig := ProcessFile(requirePath, includeDirs)
		config = config.Merge(requireConfig)
	}
	return config
}

// findFile will return the first combination of includeDirs and fileName that exists on the system
func findFile(fileName string, includeDirs []string) (string, error) {
	// first check if the fileName points to a file that we can resolve without looking at includeDirs
	file, err := os.Stat(fileName)
	if err == nil {
		return fileName, nil
	}

	// otherwise dig through our includeDirs
	for _, dir := range includeDirs {
		filePath := path.Join(dir, fileName)
		file, err = os.Stat(filePath)
		// if we can't stat that path, we move on to the next one
		if err != nil {
			continue
			// if its a regular file, we return the full path to it
		} else if file.Mode().IsRegular() {
			return filePath, nil
		}
	}
	// if we get here, then no paths exist.  Thats an error.
	return "", errors.New(fmt.Sprintf("File not found: %s", fileName))
}
