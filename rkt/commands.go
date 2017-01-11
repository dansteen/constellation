package rkt

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Pods stores a number of pods indexed by their appName
type Pods struct {
	Pods map[string]Pod
}

// GetRunningPods will get a list of running pods that are relevant to the provided project, and will return their information indexed by
// the appName
func GetRunningPods(projectName string) (Pods, error) {
	// create a type to hold our pods
	allPods := make([]Pod, 0)
	// hold our running pods for this project
	ourPods := make(map[string]Pod)

	// get all the pods
	command := strings.Split("rkt list --format=json", " ")
	listCmd := exec.Command(command[0], command[1:]...)
	output, err := listCmd.Output()
	if err != nil {
		return Pods{}, err
	}
	err = json.Unmarshal(output, &allPods)
	if err != nil {
		return Pods{}, err
	}

	// filter on the running ones
	for _, pod := range allPods {
		if pod.State == "running" {
			// and then filter again on the ones for our project
			for _, name := range pod.AppNames {
				if strings.HasPrefix(name, fmt.Sprintf("%s-", projectName)) {
					ourPods[name] = pod
				}
			}
		}
	}
	return Pods{Pods: ourPods}, nil
}

// GetAppName will generate the app name for this container/project combination
func GetAppName(projectName string, containerName string) (string, error) {
	// generate the appname as a combination of projectName and containername
	// we also need to strip out any non-alphanumeric characters or rkt will complain
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		return "", err
	}
	appName := reg.ReplaceAllString(containerName, "")
	appName = fmt.Sprintf("%s-%s", projectName, appName)
	return appName, nil
}
