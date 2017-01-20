// container stores objects related to containers
package container

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/dansteen/constellation/rkt"
	"github.com/dansteen/constellation/state"
	"github.com/dansteen/constellation/types"
	"github.com/dansteen/constellation/util"
	//"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

// Container stores all the information about a container to operate on
type Container struct {
	Name            string
	ImageHash       string
	Image           string                `json:"image"`
	Environment     map[string]string     `json:"environment"`
	Exec            string                `json:"exec"`
	StateConditions state.StateConditions `json:"state_conditions"`
	Mounts          []Mount               `json:"mounts"`
	DependsStrings  []string              `json:"depends_on"`
	DependsOn       map[string]*Container
	Ports           []*Port
}

// Init will do the inital checking of a container to make sure it's viable.  We also pull the images.
// We can also do initial container setup here if we want (though we don't right now)
func (container *Container) Init(containers map[string]*Container, volumes map[string]types.Volume) error {

	// Make sure that any mounts reference defined volumes
	for _, mount := range container.Mounts {
		if _, ok := volumes[mount.Volume]; !ok {
			return errors.New(fmt.Sprintf("Mount in %s referenced volume %s which is not defined", container.Name, mount.Volume))
		}
	}

	// make sure that any filemonitors reference paths that are mounted from the filesystem.  Otherwie the filemonitor will
	// never trigger since it runs outside of the container
	for index, condition := range container.StateConditions.FileMonitors {
		found := false
		for _, mount := range container.Mounts {
			if strings.HasPrefix(condition.File, mount.Path) {
				found = true
				// replace the prefix with the respective local path
				localPath := strings.Replace(condition.File, mount.Path, volumes[mount.Volume].Path, 1)
				container.StateConditions.FileMonitors[index] = state.FileMonitorCondition{
					File:   localPath,
					Regex:  condition.Regex,
					Status: condition.Status,
				}
			}
		}
		if found == false {
			return errors.New(fmt.Sprintf("File monitor requests a path (%s) that is not prefixed by any mount path", condition.File))
		}
	}

	// run through the dependency strings and link up the containers to DependsOn
	depends := make(map[string]*Container)
	for _, containerName := range container.DependsStrings {
		if _, ok := containers[containerName]; ok {
			depContainer := containers[containerName]
			depends[containerName] = depContainer
		} else {
			return errors.New(fmt.Sprintf("%s depends on %s which does not exist in the config.", container.Name, containerName))
		}
	}
	container.DependsOn = depends

	// pull our image
	imageHash, err := rkt.Fetch(container.Image)
	if err != nil {
		return err
	}
	container.ImageHash = imageHash

	// initialize our port list
	container.Ports = make([]*Port, 0)
	// grab our image manifest (we use the hash hear because cat-manifest doesn't like docker images
	imageManifest, err := rkt.GetImageManifest(container.ImageHash)
	if err != nil {
		log.Printf("Could not get manifest of image: %s", container.ImageHash)
		return err
	}
	// and add the ports to our container
	for _, manifestPort := range imageManifest.App.Ports {
		container.Ports = append(container.Ports, &Port{ImageAppPort: manifestPort})
	}

	return nil

}

// Run will run a container.  It will return an error message if the container fails by any of the containers StateConditions
func (container *Container) Run(configPath string, projectName string, volumes map[string]types.Volume, hostsEntries []types.HostsEntry) error {
	// set up logging for this run
	colors := util.RandomColor()
	ourColor := color.New(colors...).SprintfFunc()
	logger := log.New(os.Stdout, fmt.Sprintf("[%s] ", ourColor(container.Name)), log.LstdFlags)
	logger.Printf("Running")

	// set a result to start with
	var result error
	result = nil

	// check to see if we are not already running a container with this project and name
	// get our name
	name, err := rkt.GetAppName(projectName, container.Name)
	// get a list of running pods
	runningPods, err := rkt.GetRunningPods(projectName)
	if err != nil {
		return err
	}
	// a pod of our name is already running, we dont continue
	for runningName, _ := range runningPods.Pods {
		if runningName == name {
			logger.Printf("Using already running container %s for %s.", runningName, container.Name)
			return nil
		}
	}

	// get our command line
	commandLine, err := container.getCommandLine(projectName, runningPods, logger)
	if err != nil {
		return err
	}

	// prefix volumes
	for _, volume := range volumes {
		commandLine = append(volume.GenerateCommandLine(), commandLine...)
	}

	// prefix hostsEntries
	for _, entry := range hostsEntries {
		commandLine = append(entry.GenerateCommandLine(), commandLine...)
	}

	// prefix our port maps
	for _, entry := range container.Ports {
		// we do this as close to execution as possible to avoid conflicts
		err = entry.SetHostPort()
		if err != nil {
			return err
		}
		commandLine = append(entry.GenerateCommandLine(), commandLine...)
	}

	// prefix TODO: we want to allow settings for these
	commandLine = append(strings.Split(fmt.Sprintf("rkt run --local-config=%s --dns=host", configPath), " "), commandLine...)

	logger.Println(commandLine)
	// set up our command run
	command := exec.Command(commandLine[0], commandLine[1:]...)

	// setup our state condition results
	status := make(chan error)
	// setup our stop channel to let state conditions know they don't need to continue, we buffer as many values as we have handlers
	numHandlers := container.StateConditions.Count()
	stop := make(chan bool, numHandlers)

	// handle timeouts if set
	if container.StateConditions.Timeout != nil {
		go container.StateConditions.Timeout.Handle(status, stop, logger)
	}

	// handle log monitors if set (must happen before command is started)
	if len(container.StateConditions.FileMonitors) > 0 {
		for _, monitor := range container.StateConditions.FileMonitors {
			go func(monitor state.FileMonitorCondition, status chan error, stop chan bool, logger *log.Logger) {
				monitor.Handle(status, stop, logger)
			}(monitor, status, stop, logger)
		}
	}

	// we want to both monitor and print outputs so we do things a bit different for this Handler.  This has to go prior to
	// command.Start()
	err = container.handleOutputs(command, status, stop, logger)
	if err != nil {
		log.Fatal(err)
	}

	// start the command
	err = command.Start()
	if err != nil {
		log.Fatal(err)
	}

	// handle exit conditions if set (must happen after the command is started)
	if container.StateConditions.Exit != nil {
		go container.StateConditions.Exit.Handle(command, status, stop, logger)
	} else {
		// if we don't have an exit handler, we build a default one to fail on any exit
		exitHandler := state.ExitCondition{
			Codes:  []int{-1},
			Status: "success",
		}
		go exitHandler.Handle(command, status, stop, logger)
	}

	// we wait for one of our conditions to return if we have any
	if container.StateConditions.Count() != 0 {
		result = <-status
		// once one condition returns, we cancel the rest
		for i := 0; i < numHandlers; i++ {
			stop <- true
		}
	}

	return result
}

// handleOutputs will print the stderr and stdout of command
func (container *Container) handleOutputs(command *exec.Cmd, results chan<- error, stop <-chan bool, logger *log.Logger) error {

	// process our outputs
	// stdout
	stdout, err := command.StdoutPipe()
	// check for errors
	if err != nil {
		return errors.New("Could not connect to stdout")
	}
	outScanner := bufio.NewScanner(stdout)
	go container.handleOutput(outScanner, "STDOUT", results, stop, logger)

	// stderr
	stderr, err := command.StderrPipe()
	// check for errors
	if err != nil {
		return errors.New("Could not connect to stderr")
	}
	errScanner := bufio.NewScanner(stderr)
	go container.handleOutput(errScanner, "STDERR", results, stop, logger)
	return nil
}

// handleOutput does the heavy lifting for printOutputs.  Source is the source the log is coming from
// this also activates the state condition handler for outputs since we can only tap into the outputs a single time
func (container *Container) handleOutput(scanner *bufio.Scanner, source string, results chan<- error, stop <-chan bool, logger *log.Logger) {
	// we print app messages a different color so they stand out
	appMessage := color.New(color.FgWhite, color.BgBlack).SprintFunc()

	// see if we need to run an output condition for this
	conditions := make([]*state.OutputCondition, 0)
	for _, outputCondition := range container.StateConditions.Outputs {
		if outputCondition.Source == source {
			logger.Printf("Monitoring %s for %+v\n", outputCondition.Source, outputCondition.Regex)
			conditions = append(conditions, &outputCondition)
		}
	}

	// first print our output
	for {
		// check if we still need to handle state conditions
		select {
		case <-stop:
			logger.Printf("Cancelled Monitoring %s\n", source)
			conditions = make([]*state.OutputCondition, 0)
		default:
			moreContent := scanner.Scan()
			// if we hit an error or eof we are done
			if !moreContent {
				logger.Println(scanner.Err())
				return
			}
			logger.Printf("%s", appMessage(scanner.Text()))
			// if we need to handle the content
			for _, condition := range conditions {
				condition.Handle(scanner.Text(), results, stop, logger)
			}
		}
	}
}

// getCommandLine will generate rkt cli commands for this container
func (container *Container) getCommandLine(projectName string, runningPods rkt.Pods, logger *log.Logger) ([]string, error) {
	// generate the different components
	command := make([]string, 0)

	// get the appName
	appName, err := rkt.GetAppName(projectName, container.Name)
	if err != nil {
		return command, err
	}
	appNameLine := fmt.Sprintf("--name=%s", appName)

	// generate environment strings
	envArray := make([]string, 0)
	for varName, varValue := range container.Environment {
		envArray = append(envArray, fmt.Sprintf("--environment=%s=%s", varName, varValue))
	}

	// exec string
	execArray := make([]string, 0)
	if container.Exec != "" {
		exec_parts := util.ShellSplit(container.Exec)

		if len(exec_parts) > 0 {
			// first prime our array
			execArray = append(execArray, "--exec")
			// split our string into parts
			execArray = append(execArray, exec_parts[0])
		}
		// if there is more than one part the rkt command requires that other compoments come after a double hyphen
		if len(exec_parts) > 1 {
			execArray = append(execArray, "--")
			execArray = append(execArray, exec_parts[1:]...)
		}
	}

	// mount strings
	mountArray := make([]string, 0)
	for _, mount := range container.Mounts {
		mountArray = append(mountArray, mount.GenerateCommandLine()...)
	}

	depIPMap, err := container.GetDepChainIPs(projectName, runningPods, logger)
	if err != nil {
		return command, err
	}

	hostsArray := make([]string, 0)
	for name, IPs := range depIPMap {
		for _, IP := range IPs {
			hostsArray = append(hostsArray, fmt.Sprintf("--hosts-entry=%s=%s", IP, name))
		}
	}

	// create the hostname
	hostnameLine := fmt.Sprintf("--hostname=%s", container.Name)

	// combine our command parts
	command = append(command, container.Image)
	command = append(command, hostnameLine)
	command = append(command, envArray...)
	command = append(command, mountArray...)
	command = append(command, hostsArray...)
	command = append(command, appNameLine)
	command = append(command, execArray...)

	return command, nil
}

// getDependencyChainIPs will return a map of container name=>IP of each dependency of the container and each of their dependencies
func (container *Container) GetDepChainIPs(projectName string, runningPods rkt.Pods, logger *log.Logger) (map[string][]string, error) {
	// store our ips and names
	depIPMap := make(map[string][]string)
	// run through the dependencies
	for name, depContainer := range container.DependsOn {
		// create our ip array
		depIPMap[name] = make([]string, 0)
		// get the appName for this depend
		depAppName, err := rkt.GetAppName(projectName, name)
		if err != nil {
			return depIPMap, err
		}

		// check if there is a running pod, and if it is, grab the ip
		if _, ok := runningPods.Pods[depAppName]; ok {
			for _, network := range runningPods.Pods[depAppName].Networks {
				depIPMap[name] = append(depIPMap[name], network.IP)
			}
		} else {
			// if there is not, check the pod to make sure that it was allowed to exit
			if container.DependsOn[name].StateConditions.Exit != nil && container.DependsOn[name].StateConditions.Exit.Status == "success" {
				logger.Printf("Required dependency %s is not running.  Looks like it is allowed to exit so we are ignoring. \n", name)
			} else {
				logger.Printf("Required dependency %s is not running.  No valid exit state.  Failing. \n", name)
				return depIPMap, errors.New(fmt.Sprintf("Required dependency %s is not running", name))
			}
		}
		// run on each dependency so we get a full set of heirachical IPs
		depDepIPMap, err := depContainer.GetDepChainIPs(projectName, runningPods, logger)
		if err != nil {
			return depIPMap, err
		}
		// merge our maps
		for name, value := range depDepIPMap {
			if _, ok := depIPMap[name]; ok {
				logger.Printf("Already have IP record for %s.  Ignoring duplicate.\n", name)
			} else {
				depIPMap[name] = value
			}
		}
	}
	return depIPMap, nil
}
