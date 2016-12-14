package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/dansteen/constellation/types"
	"github.com/dansteen/constellation/util"

	"gopkg.in/yaml.v2"
)

// ProcessFile will process config files for constellation, and return an array of Container objects
// required files are processed as well.  Accepts a filepath and an array of directories to search for
// files that are included via the 'require' stanza
func ProcessFile(filePath string, includeDirs []string) []types.Container {

	// setup a map to hold our config
	var config map[string]interface{}
	// an array for our containers
	containers := make([]types.Container, 0)

	// read in the file provided
	data, err := ioutil.ReadFile(filePath)
	util.Check(err)
	err = yaml.Unmarshal(data, &config)
	util.Check(err)

	// start with our base possible stanzas
	for key, data := range config {
		switch key {
		case "require":
			// make sure the value is a string
			switch data.(type) {
			case string:
				// do nothing
				_ = 1
			default:
				util.Check(errors.New(fmt.Sprintf("Invalid value for 'require' stanza in %s.  Needs a String type", filePath)))
			}
			filePath, err := findFile(data.(string), includeDirs)
			util.Check(err)
			// get containers from the requires and add them to our list
			requireContainers := ProcessFile(filePath, make([]string, 0))
			containers = append(containers, requireContainers...)
		case "containers":
			log.Println("containers")
		default:
			log.Printf("invalid stanza: %s.", key)
		}
	}
	return containers
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
