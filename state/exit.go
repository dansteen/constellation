package state

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"syscall"
)

type ExitCondition struct {
	Codes  []int  `json:codes`
	Status string `json:status`
}

func (cond *ExitCondition) Handle(command *exec.Cmd, results chan<- error, stop <-chan bool, logger *log.Logger) {
	logger.Printf("Waiting for Exit %+v\n", cond.Codes)
	// report the results of our wait
	waitResult := make(chan error)
	// start a command wait
	go func(command *exec.Cmd, result chan<- error) {
		err := command.Wait()
		result <- err
	}(command, waitResult)

	// wait for the command to exit and grab the exit code or listen for a stop command
	var exitCode int
	select {
	case err := <-waitResult:
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				// The program has exited with an exit code != 0
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
				}
			}
		} else {
			exitCode = 0
		}
		logger.Printf("Received Exit Code: %d\n", exitCode)
	case <-stop:
		logger.Println("Waiting for Exit Cancelled")
		return
	}

	// once we have the exit code, check if it's in our list
	for _, code := range cond.Codes {
		if code == exitCode {
			switch cond.Status {
			case "success":
				results <- nil
				return
			case "failure":
				results <- errors.New(fmt.Sprintf("Exit code %d specified as failure\n", exitCode))
				return
			}
		}
	}
	// if it's not, we do the opposite of the Status
	switch cond.Status {
	case "success":
		results <- errors.New(fmt.Sprintf("Exit code %d specified as failure\n", exitCode))
		return
	case "failure":
		results <- nil
		return
	}
}
