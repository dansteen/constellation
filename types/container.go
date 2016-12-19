// types stores objects of various types and the functions to interact with them
package types

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/dansteen/constellation/state"
	"github.com/dansteen/constellation/util"
)

// Container stores all the information about a container to operate on
type Container struct {
	Name            string
	Image           string            `json:image`
	Environment     map[string]string `json:environment`
	Exec            []string          `json:exec`
	StateConditions []StateCondition  `json:state_conditions`
	Mounts          []Mount           `json:mounts`
	DependsOn       []string          `json:depends_on`
}

func NewContainer(name string) Container {
	return Container{
		Name: name,
	}
}

// SetImageFromConfig will set the image values from config passed in via YAML.  It will check the values to make
// sure they are valid
func (container *Container) SetImageFromConfig(config interface{}) error {
	// make sure the type is correct
	switch config.(type) {
	case string:
		container.Image = config.(string)
	default:
		return errors.New(fmt.Sprintf("image stanza requires a string.  Got %s", reflect.TypeOf(config)))
	}
	return nil
}

// SetExecFromConfig will set the image values from config passed in via YAML.  It will check the values to make
// sure they are valid
func (container *Container) SetExecFromConfig(config interface{}) error {
	// make sure the type is correct
	switch config.(type) {
	case string:
		container.Exec = util.ShellSplit(config.(string))
	default:
		return errors.New(fmt.Sprintf("exec stanza requires a string.  Got %s", reflect.TypeOf(config)))
	}
	return nil
}

// SetEnvironmentFromConfig will set the image values from config passed in via YAML.  It will check the values to make
// sure they are valid
func (container *Container) SetEnvironmentFromConfig(config interface{}) error {
	// initialize our enviroment map
	container.Environment = make(map[string]string)
	// make sure the type is correct in general
	switch config.(type) {
	case map[interface{}]interface{}:
		// run through each entry
		for envVar, value := range config.(map[interface{}]interface{}) {
			// make sure each entry is a string
			switch envVar.(type) {
			case string:
				// and each value
				switch value.(type) {
				case string:
					container.Environment[envVar.(string)] = value.(string)
				default:
					return errors.New(fmt.Sprintf("environment values require String.  Got %s for %s", reflect.TypeOf(value), envVar))
				}
			default:
				return errors.New(fmt.Sprintf("environment Vars Need to be a string.  Got %s", reflect.TypeOf(value)))
			}
		}
	default:
		return errors.New(fmt.Sprintf("environment stanza requires a Hash.  Got %s", reflect.TypeOf(config)))
	}
	return nil
}

// SetMountsFromConfig will set the mount values from config passed in via YAML. It will check to make sure values are valid
func (container *Container) SetMountsFromConfig(config interface{}) error {
	// initialize our mount array
	container.Mounts = make([]Mount, 0)
	// make sure the type is correct in general
	switch config.(type) {
	case []interface{}:
		for index, mountSpec := range config.([]interface{}) {
			// make sure a string has been specified
			switch mountSpec.(type) {
			case string:
				mountArray := strings.SplitN(mountSpec.(string), ":", 1)
				if len(mountArray) != 2 {
					return errors.New(fmt.Sprintf("mount strings need to be in the format <volume>:<path>"))
				}
				container.Mounts = append(container.Mounts, NewMount(mountArray[0], mountArray[1]))
			default:
				return errors.New(fmt.Sprintf("mounts need to be strings. Got %s for %b", reflect.TypeOf(mountSpec), index))
			}
		}
	default:
		return errors.New(fmt.Sprintf("mounts need to be an array of strings.  Got %s", reflect.TypeOf(config)))
	}
	return nil
}

// SetDependsOnFromConfig will set the depends_on values from config passed in via YAML. It will check to make sure values are valid
func (container *Container) SetDependsOnFromConfig(config interface{}) error {
	// initialize our array
	container.DependsOn = make([]string, 0)
	// make sure the type is correct in general
	switch config.(type) {
	case []interface{}:
		for index, dependency := range config.([]interface{}) {
			// make sure a string has been specified
			switch dependency.(type) {
			case string:
				container.DependsOn = append(container.DependsOn, dependency.(string))
			default:
				return errors.New(fmt.Sprintf("depends_on need to be a string. Got %s for %b", reflect.TypeOf(dependency), index))
			}
		}
	default:
		return errors.New(fmt.Sprintf("depends_on needs to be an array of strings.  Got %s", reflect.TypeOf(config)))
	}
	return nil
}

// SetStateConditionsFromConfig will set the depends_on values from config passed in via YAML. It will check to make sure values are valid
func (container *Container) SetStateConditionsFromConfig(config interface{}) error {
	// initialize our array
	container.StateConditions = make([]StateCondition, 0)
	// make sure the type is correct in general
	switch config.(type) {
	case map[interface{}]interface{}:
		for condType, condition := range config.(map[interface{}]interface{}) {
			// make sure a string has been specified
			switch condType {
			case "exit":
				// get our exitCondition
				exitCondition, err := state.NewExitConditionFromConfig(condition)
				if err != nil {
					return errors.New(fmt.Sprintf("%s in state_conditions", err))
				}
				container.StateConditions = append(container.StateConditions, exitCondition)
			case "filemonitor":
				// get our fileMonitor Conditions
				fileMonitorCondition, err := state.NewFileMonitorConditionFromConfig(condition)
				if err != nil {
					return errors.New(fmt.Sprintf("%s in state_conditions", err))
				}
				container.StateConditions = append(container.StateConditions, fileMonitorCondition)
			case "timeout":
				// get our timeoutConditions
				timeoutCondition, err := state.NewTimeoutConditionFromConfig(condition)
				if err != nil {
					return errors.New(fmt.Sprintf("%s in state_conditions", err))
				}
				container.StateConditions = append(container.StateConditions, timeoutCondition)
			default:
				return errors.New(fmt.Sprintf("invalid state_condition %s", condType))
			}
		}
	default:
		return errors.New(fmt.Sprintf("state_conditions needs to be a Hash.  Got %s", reflect.TypeOf(config)))
	}
	return nil
}
