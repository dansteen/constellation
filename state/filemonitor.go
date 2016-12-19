package state

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

type FileMonitorCondition struct {
	File   string        `json:file`
	Regex  regexp.Regexp `json:regex`
	Status string        `json:status`
}

// NewFileMonitorConditionFromConfig will create an exist condition from the provided config
func NewFileMonitorConditionFromConfig(config interface{}) (FileMonitorCondition, error) {
	// create a new file monitor condition and prime it
	condition := FileMonitorCondition{}

	// make sure the value type is correct in general
	switch config.(type) {
	case map[interface{}][]interface{}:
		for stanza, value := range config.(map[string]interface{}) {
			// make sure a string has been specified
			switch stanza {
			case "codes":
				switch value.(type) {
				case []int:
					condition.File = value.(string)
				default:
					return FileMonitorCondition{}, errors.New(fmt.Sprintf("codes must to be an array of int. Got %s for exit", reflect.TypeOf(value)))
				}
			case "status":
				switch value.(type) {
				case string:
					switch value {
					case "success", "failure":
						condition.Status = value.(string)
					default:
						return FileMonitorCondition{}, errors.New(fmt.Sprintf("status needs to be one of 'success' or 'failure'. Got %s for exit", value))
					}
				default:
					return FileMonitorCondition{}, errors.New(fmt.Sprintf("exit expects stanzas 'codes' or 'status'. Got %s", stanza))
				}
			}
		}
	default:
		return FileMonitorCondition{}, errors.New(fmt.Sprintf("Needs to be a Hash type for exit.  Got %s", reflect.TypeOf(config)))
	}
	return condition, nil
}
