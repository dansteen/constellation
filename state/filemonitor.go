package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/hpcloud/tail"
)

type FileMonitorCondition struct {
	File   string         `json:"file"`
	Regex  *regexp.Regexp `json:"regex"`
	Status string         `json:"status"`
}

func (monitor *FileMonitorCondition) UnmarshalJSON(b []byte) error {
	// create a string version of our monitor
	type StringMonitor struct {
		File   string `json:"file"`
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
	// then create a new FileMonitorCondition with our new values
	monitor.File = stringMonitor.File
	monitor.Status = stringMonitor.Status
	monitor.Regex = regex
	return nil
}

// handleFile handler for actual files
func (monitor *FileMonitorCondition) Handle(results chan<- error, stop <-chan bool, logger *log.Logger) {
	logger.Printf("Monitoring %s for %s\n", monitor.File, monitor.Regex.String())

	// tail our file
	tail, err := tail.TailFile(monitor.File, tail.Config{Follow: true, ReOpen: true, MustExist: false, Logger: tail.DiscardingLogger})
	if err != nil {
		log.Fatal(err)
	}

	// then check for our regexs
	for line := range tail.Lines {
		// if we get signalled that we are done we also exit
		select {
		case <-stop:
			fmt.Printf("Cancelled Monitoring %s for %s\n", monitor.File, monitor.Regex.String())
			return
		default:
			if monitor.Regex.Match([]byte(line.Text)) == true {
				logger.Printf("%s matched %s.\n", monitor.File, monitor.Regex.String())
				if monitor.Status == "success" {
					results <- nil
				}
				if monitor.Status == "failure" {
					results <- errors.New(fmt.Sprintf("%s matched %s. Specified as failure\n", monitor.File, monitor.Regex.String()))
				}
				return
			}
		}
	}
}
