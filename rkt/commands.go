package rkt

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// Pods stores a number of pods indexed by their appName
type Pods struct {
	Pods map[string]Pod
}

// GetAllPods will return a list of all pods in rkt
func GetAllPods(projectName string) (Pods, error) {
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

	// filter on the ones for our project
	for _, pod := range allPods {
		for _, name := range pod.AppNames {
			if strings.HasPrefix(name, fmt.Sprintf("%s-", projectName)) {
				ourPods[name] = pod
			}
		}
	}
	return Pods{Pods: ourPods}, nil
}

// GetRunningPods will get a list of running pods that are relevant to the provided project, and will return their information indexed by
// the appName
func GetRunningPods(projectName string) (Pods, error) {
	// grab our pods
	pods, err := GetAllPods(projectName)
	if err != nil {
		return Pods{}, err
	}
	// strip out our running pods
	for name, pod := range pods.Pods {
		if pod.State != "running" {
			delete(pods.Pods, name)
		}
	}
	return pods, nil
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

// Fetch will fetch a rkt image and return the image hash
func Fetch(image string) (string, error) {
	log.Printf("Fetching image: %s", image)
	// fetch our pod
	command := strings.Split(fmt.Sprintf("rkt fetch --insecure-options=all-fetch --trust-keys-from-https=true %s", image), " ")
	listCmd := exec.Command(command[0], command[1:]...)
	output, err := listCmd.Output()
	// if there is an error, print the output
	if err != nil {
		log.Printf("%s", output)
	}
	return strings.TrimSpace(string(output)), err
}

// GetImageManifest will get the manifest for an image that has been fetched
// it returns the manifest mapped to a generic interface
func GetImageManifest(image string) (*ImageManifest, error) {
	// setup an object to hold our ImageManifest
	imageManifest := ImageManifest{}
	// grab our manifest in json
	command := strings.Split(fmt.Sprintf("rkt image cat-manifest %s", image), " ")
	listCmd := exec.Command(command[0], command[1:]...)
	output, err := listCmd.Output()
	// if there is an error, print the output
	if err != nil {
		log.Printf("%s", output)
		return &imageManifest, err
	}
	// otherwise we unmarshal and return the manifest
	err = yaml.Unmarshal(output, &imageManifest)
	if err != nil {
		return &imageManifest, err
	}
	return &imageManifest, nil
}
