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
			// ensure that the value is what we expect
			switch data.(type) {
			// if it is, we process
			case string:
				filePath, err := findFile(data.(string), includeDirs)
				util.Check(err)
				// get containers from the requires and add them to our list
				requireContainers := ProcessFile(filePath, make([]string, 0))
				containers = append(containers, requireContainers...)
				// otherwise we error
			default:
				util.Check(errors.New(fmt.Sprintf("Invalid value for 'require' stanza in %s.  Needs a String type", filePath)))
			}
		case "containers":
			// make sure the value is a map
			switch data.(type) {
			case map[interface{}]interface{}:
				// run through each container stanza
				for name, containerData := range data.(map[interface{}]interface{}) {
					// make sure the container data is valid
					newContainer, err := processContainerConfig(containerData, name.(string))
					if err != nil {
						util.Check(errors.New(fmt.Sprintf("%s in %s", err.Error(), filePath)))
					}
					containers = append(containers, newContainer)
				}
			default:
				util.Check(errors.New(fmt.Sprintf("Invalid value for 'containers' stanza in %s.  Needs a Hash type", filePath)))
			}
		default:
			log.Printf("invalid stanza: %s.", key)
		}
	}
	log.Printf("%+v\n", containers)
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

// processContainerConfig will generate a container from the provided config
func processContainerConfig(config interface{}, name string) (types.Container, error) {
	// create our new container
	container := types.NewContainer(name)
	// make sure the config is valid
	switch config.(type) {
	// if it is
	case map[interface{}]interface{}:
		// create a value to hold our error
		var err error
		// run through our config and handle each stanza
		for stanza, data := range config.(map[interface{}]interface{}) {
			switch stanza {
			case "image":
				if err = container.SetImageFromConfig(data); err != nil {
					break
				}
			case "exec":
				if err = container.SetExecFromConfig(data); err != nil {
					break
				}
			case "mounts":
				if err = container.SetMountsFromConfig(data); err != nil {
					break
				}
			case "state_conditions":
				if err = container.SetStateConditionsFromConfig(data); err != nil {
					break
				}
			case "depends_on":
				if err = container.SetDependsOnFromConfig(data); err != nil {
					break
				}
			case "environment":
				if err = container.SetEnvironmentFromConfig(data); err != nil {
					break
				}
			default:
				err = errors.New(fmt.Sprintf("Invalid stanza %s", stanza))
				break
			}
		}
		if err != nil {
			return types.Container{}, errors.New(fmt.Sprintf("%s in %s", err.Error(), name))
		}
		// if we don't recognize the type we error out
	default:
		return types.Container{}, errors.New(fmt.Sprintf("Needs a Hash type for container %s", name))
	}
	return container, nil
}
