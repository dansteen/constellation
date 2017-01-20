package state

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type TimeoutCondition struct {
	Duration int64  `json:duration`
	Status   string `json:status`
}

func (cond *TimeoutCondition) Handle(results chan<- error, stop <-chan bool, logger *log.Logger) {
	logger.Printf("Waiting for timeout: %ds", cond.Duration)
	// start a timeout
	timeout := make(chan bool)
	go func() {
		time.Sleep(time.Second * time.Duration(cond.Duration))
		select {
		case timeout <- true:
		case <-stop:
			return
		}
	}()
	// wait for our timeout or the done channel
	select {
	case <-timeout:
		logger.Println("Hit Timeout")
	case <-stop:
		return
	}

	// depending on what the status is set to be we publish our result
	switch cond.Status {
	case "success":
		results <- nil
	case "failure":
		results <- errors.New(fmt.Sprintf("Hit Timeout. Specified as failure."))
	}
}
