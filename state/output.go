package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
)

type OutputCondition struct {
	Source string         `json:"source"`
	Regex  *regexp.Regexp `json:"regex"`
	Status string         `json:"status"`
}

func (monitor *OutputCondition) UnmarshalJSON(b []byte) error {
	// create a string version of our monitor
	type StringMonitor struct {
		Source string `json:"source"`
		String string `json:"regex"`
		Status string `json:"status"`
	}
	var stringMonitor StringMonitor
	// unmarshal our items into it
	err := json.Unmarshal(b, &stringMonitor)
	if err != nil {
		return err
	}

	// then convert ouf String to a regex
	regex, err := regexp.Compile(stringMonitor.String)
	if err != nil {
		return err
	}
	// then create a new OutputCondition with our new values
	monitor.Source = stringMonitor.Source
	monitor.Status = stringMonitor.Status
	monitor.Regex = regex
	return nil
}

// Handle will handle the output condition.  This handler is different than the ones for the other conditions in/
// that it expects a different process to manage the streams.
func (monitor *OutputCondition) Handle(logLine string, results chan<- error, stop <-chan bool, logger *log.Logger) {
	// then check for our regexs
	// once we find a match we are done
	if monitor.Regex.Match([]byte(logLine)) == true {
		logger.Printf("%s matched %s.\n", monitor.Source, monitor.Regex.String())
		if monitor.Status == "success" {
			results <- nil
		}
		if monitor.Status == "failure" {
			results <- errors.New(fmt.Sprintf("%s matched %s. Specified as failure\n", monitor.Source, monitor.Regex.String()))
		}
		return
	}
}
