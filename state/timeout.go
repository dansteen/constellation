package state

import (
	"errors"
	"fmt"
	"reflect"
)

type TimeoutCondition struct {
	Duration int    `json:duration`
	Status   string `json:status`
}

// NewTimeoutConditionFromConfig will create an exist condition from the provided config
func NewTimeoutConditionFromConfig(config interface{}) (TimeoutCondition, error) {
	// create a new exit condition and prime it
	condition := TimeoutCondition{}

	// make sure the value type is correct in general
	switch config.(type) {
	case map[interface{}]interface{}:
		for stanza, value := range config.(map[string]interface{}) {
			// make sure a string has been specified
			switch stanza {
			case "duration":
				switch value.(type) {
				case int:
					condition.Duration = value.(int)
				default:
					return TimeoutCondition{}, errors.New(fmt.Sprintf("duration must to be a number of seconds. Got %s for timeout", reflect.TypeOf(value)))
				}
			case "status":
				switch value.(type) {
				case string:
					switch value {
					case "success", "failure":
						condition.Status = value.(string)
					default:
						return TimeoutCondition{}, errors.New(fmt.Sprintf("status needs to be one of 'success' or 'failure'. Got %s for timeout", value))
					}
				default:
					return TimeoutCondition{}, errors.New(fmt.Sprintf("timeout expects stanzas 'duration' or 'status'. Got %s", stanza))
				}
			}
		}
	default:
		return TimeoutCondition{}, errors.New(fmt.Sprintf("Needs to be a Hash type for timeout.  Got %s", reflect.TypeOf(config)))
	}
	return condition, nil
}
