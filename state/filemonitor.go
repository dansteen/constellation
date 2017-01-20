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

	// tail our file. We seek to the end first
	tail, err := tail.TailFile(monitor.File, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: 2,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-stop:
			tail.Stop()
			return
		case line := <-tail.Lines:
			if monitor.Regex.Match([]byte(line.Text)) == true {
				if monitor.Status == "success" {
					logger.Printf("Matched %s to %s. Success.\n", monitor.File, monitor.Regex.String())
					results <- nil
				}
				if monitor.Status == "failure" {
					results <- errors.New(fmt.Sprintf("Matched %s to %s. Specified as failure\n", monitor.File, monitor.Regex.String()))
				}
				// stop tailing
				tail.Stop()
				return
			}
		}
	}
}
