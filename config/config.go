package config

import (
	"errors"
	"fmt"

	"encoding/json"

	"github.com/dansteen/constellation/container"
	"github.com/dansteen/constellation/types"
	"github.com/twmb/algoimpl/go/graph"
)

// Config holds the config in a file
type Config struct {
	Containers map[string]*container.Container `json:"containers"`
	Requires   []string                        `json:"require"`
	Volumes    map[string]types.Volume         `json"volumes"`
}

// UnmarshalJSON
func (config *Config) UnmarshalJSON(b []byte) error {
	// create a struct to unmarshal into
	type TempConfig struct {
		Containers map[string]*container.Container `json:"containers"`
		Requires   []string                        `json:"require"`
		Volumes    map[string]types.Volume         `json"volumes"`
	}
	var tempConfig TempConfig
	// unmarshal our items into the container
	err := json.Unmarshal(b, &tempConfig)
	if err != nil {
		return err
	}

	// add the names into each container
	for name, container := range tempConfig.Containers {
		// we have to do this a bit round-aboutly do to https://github.com/golang/go/issues/3117
		container.Name = name
		tempConfig.Containers[name] = container
	}
	// add the names of the volumes
	for name, volume := range tempConfig.Volumes {
		// we have to do this a bit round-aboutly do to https://github.com/golang/go/issues/3117
		volume.Name = name
		tempConfig.Volumes[name] = volume
	}

	// set the values in our config
	config.Containers = tempConfig.Containers
	config.Requires = tempConfig.Requires
	config.Volumes = tempConfig.Volumes
	return nil
}

// Merge will merge two configs, but not overwrite existing data
func (config Config) Merge(newConfig Config) Config {
	// run through the new continers
	for name, container := range newConfig.Containers {
		// make sure we arent overwriting existing values
		if _, ok := config.Containers[name]; ok {
			fmt.Printf("Already seen container %s.  Ignoring second instance\n", name)
		} else {
			// do the merge
			config.Containers[name] = container
		}
	}

	// run through new volumes
	for name, volume := range newConfig.Volumes {
		// make sure we arent overwriting existing values
		if _, ok := config.Volumes[name]; ok {
			fmt.Printf("Already seen Volume %s.  Ignoring second instance\n", name)
		} else {
			// do the merge
			config.Volumes[name] = volume
		}
	}

	// merge the requires just for completion
	config.Requires = append(config.Requires, newConfig.Requires...)
	return config
}

// DependencyOrder build a sorted list of containers based on each containers dependencies.
// We use a topological sort for this.  Circular dependencies result in an error
func (config *Config) DependencyOrder() ([]string, error) {
	// create a new graph
	ourGraph := graph.New(graph.Directed)
	// a place to store our nodes
	nodes := make(map[string]graph.Node)

	// add in nodes for each of our containers
	for name, _ := range config.Containers {
		nodes[name] = ourGraph.MakeNode()
		// hook the data back into the graph (not strictly required)
		*nodes[name].Value = name
	}

	// add in our edges
	for name, container := range config.Containers {
		for _, dep := range container.DependsStrings {
			// make sure the dependency exists
			if _, found := nodes[dep]; !found {
				return make([]string, 0), errors.New(fmt.Sprintf("Container %v depends on %v which is not included in the config\n", name, dep))
			}
			// add in an edge for this dependency
			ourGraph.MakeEdge(nodes[dep], nodes[name])
		}
	}

	// do our sort
	sorted := ourGraph.TopologicalSort()

	// generate an array of names
	orderedContainerNames := make([]string, 0)
	for _, node := range sorted {
		orderedContainerNames = append(orderedContainerNames, (*node.Value).(string))
	}

	return orderedContainerNames, nil
}
