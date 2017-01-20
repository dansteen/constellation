// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/dansteen/constellation/config"
	"github.com/dansteen/constellation/types"
	"github.com/dansteen/constellation/util"
	//"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the containers specified in the supplied config file",
	Long:  ``,
	Run:   run,
}

func init() {
	RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func run(cmd *cobra.Command, args []string) {
	log.Println("run called")
	BaseInit()
	// grab some config items
	projectName := viper.GetString("projectName")
	netConfigPath := viper.GetString("netConfigPath")
	constellationFile := viper.GetString("constellationFile")
	includeDirs := viper.GetStringSlice("includeDirs")
	imageOverrides := viper.GetStringSlice("imageOverrides")
	volumeOverrides := viper.GetStringSlice("volumeOverrides")
	hostsEntries := viper.GetStringSlice("hostsEntries")

	// generate our network file
	netConfigFile := fmt.Sprintf("%s/net.d/%s.conf", netConfigPath, projectName)
	// first check if we already have a config file for this project
	if _, err := os.Stat(netConfigFile); os.IsNotExist(err) {
		network, err := types.NewNetworkConfig(projectName)
		util.Check(err)
		// convert our config to json
		networkJSON, err := json.MarshalIndent(network, "", "  ")
		// make our config folder
		util.Check(os.MkdirAll(fmt.Sprintf("%s/net.d", netConfigPath), 0755))
		// write our network files
		util.Check(ioutil.WriteFile(netConfigFile, networkJSON, 0644))
	} else {
		log.Printf("Using config from previous project run: %s", netConfigPath)
	}

	// process our configs
	configData := config.ProcessFile(constellationFile, includeDirs)

	// handle image overrides passed into the command line.
	imageRE := regexp.MustCompile("(:^|[^/]*/)?([^:]*):?(.*)")
	for _, override := range imageOverrides {
		// break the override into parts
		overrideParts := imageRE.FindStringSubmatch(override)
		overrideSource := overrideParts[1]
		overrideName := overrideParts[2]
		overrideVersion := overrideParts[3]
		for name, container := range configData.Containers {
			// break the image name into parts
			containerParts := imageRE.FindStringSubmatch(container.Image)
			containerSource := containerParts[1]
			containerName := containerParts[2]
			// if it matches
			if containerName == overrideName {
				newImage := make([]string, 0)
				// generate our new image line
				if len(overrideSource) == 0 {
					newImage = append(newImage, containerSource)
				} else {
					newImage = append(newImage, overrideSource)
				}
				newImage = append(newImage, overrideName)
				newImage = append(newImage, ":", overrideVersion)
				container.Image = strings.Join(newImage, "")
				configData.Containers[name] = container
			}
		}
	}

	// handle volume overrides passed into the command line
	for _, override := range volumeOverrides {
		overrideParts := strings.SplitN(override, ":", 2)
		volName := overrideParts[0]
		volPath := overrideParts[1]
		for name, volumes := range configData.Volumes {
			if name == volName {
				volumes.Path = volPath
				configData.Volumes[name] = volumes
			}
		}
	}

	// handle hostsEntries passed into the command line
	customHosts := make([]types.HostsEntry, 0)
	for _, entry := range hostsEntries {
		host, err := types.HostsEntryFromString(entry)
		util.Check(err)
		customHosts = append(customHosts, host)
	}

	// initialize the containers
	for _, container := range configData.Containers {
		util.Check(container.Init(configData.Containers, configData.Volumes))
	}

	// make sure to create our log volumes
	for _, volume := range configData.Volumes {
		util.Check(volume.CreateDir())
	}

	// determin the order we need to execute in to satisfy dependencies
	order, err := configData.DependencyOrder()
	util.Check(err)
	//spew.Dump(configData)

	// print out the execution order
	log.Println("Will start containers in the following order:")
	for _, name := range order {
		log.Printf("\t%s\n", name)
	}

	for _, containerName := range order {
		// grab our container
		container := configData.Containers[containerName]
		// run our container
		err := container.Run(netConfigPath, projectName, configData.Volumes, customHosts)
		util.Check(err)
	}

	// after we have brought everything up we print out connection information
	for name, container := range configData.Containers {
		// we only print out infomration for containers that are not expected to exit
		if container.StateConditions.Exit == nil || container.StateConditions.Exit.Status != "success" {
			for _, port := range container.Ports {
				// TODO: get the address of the default gw interafce and print it here
				fmt.Printf("%s/%s --> %s:%d\n", name, port.Name, "address", port.HostPort)
			}
		}
	}
}
