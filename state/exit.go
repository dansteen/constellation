package state

import (
	"errors"
	"fmt"
	"reflect"
)

type ExitCondition struct {
	Codes  []int  `json:codes`
	Status string `json:status`
}

// NewExitConditionFromConfig will create an exist condition from the provided config
func NewExitConditionFromConfig(config interface{}) (ExitCondition, error) {
	// create a new exit condition and prime it
	condition := ExitCondition{
		Codes: make([]int, 0),
	}

	// make sure the value type is correct in general
	switch config.(type) {
	case map[interface{}]interface{}:
		for stanza, value := range config.(map[string]interface{}) {
			// make sure a string has been specified
			switch stanza {
			case "codes":
				switch value.(type) {
				case []int:
					condition.Codes = value.([]int)
				default:
					return ExitCondition{}, errors.New(fmt.Sprintf("codes must to be an array of int. Got %s for exit", reflect.TypeOf(value)))
				}
			case "status":
				switch value.(type) {
				case string:
					switch value {
					case "success", "failure":
						condition.Status = value.(string)
					default:
						return ExitCondition{}, errors.New(fmt.Sprintf("status needs to be one of 'success' or 'failure'. Got %s for exit", value))
					}
				default:
					return ExitCondition{}, errors.New(fmt.Sprintf("exit expects stanzas 'codes' or 'status'. Got %s", stanza))
				}
			}
		}
	default:
		return ExitCondition{}, errors.New(fmt.Sprintf("Needs to be a Hash type for exit.  Got %s", reflect.TypeOf(config)))
	}
	return condition, nil
}
