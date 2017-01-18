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
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	//"github.com/davecgh/go-spew/spew"
	"github.com/dansteen/constellation/rkt"
	"github.com/dansteen/constellation/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Stop the containers specified in the supplied config file",
	Long:  ``,
	Run:   clean,
}

func init() {
	RootCmd.AddCommand(cleanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cleanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only clean when this command
	// is called directly, e.g.:
	// cleanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func clean(cmd *cobra.Command, args []string) {
	fmt.Println("clean called")
	BaseInit()
	// get some config items
	projectName := viper.GetString("projectName")
	netConfigPath := viper.GetString("netConfigPath")

	// get our running containers for this project
	allPods, err := rkt.GetAllPods(projectName)
	util.Check(err)

	// run through and clean them
	for name, pod := range allPods.Pods {
		log.Println(name)

		// stop the container
		command := strings.Split(fmt.Sprintf("rkt stop --force %s", pod.Name), " ")
		log.Printf("Running: %+v", command)
		cleanCmd := exec.Command(command[0], command[1:]...)
		output, err := cleanCmd.CombinedOutput()
		log.Printf("%s", output)
		util.Check(err)

		// delete the container
		command = strings.Split(fmt.Sprintf("rkt rm %s", pod.Name), " ")
		log.Printf("Running: %+v", command)
		cleanCmd = exec.Command(command[0], command[1:]...)
		output, err = cleanCmd.CombinedOutput()
		log.Printf("%s", output)
		util.Check(err)

		log.Printf("Stopped and Removed %s", name)
	}

	// remove the network config
	log.Println("Removing config files")
	err = os.RemoveAll(netConfigPath)
	util.Check(err)
}
