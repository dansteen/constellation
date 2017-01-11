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
	"os/exec"
	"strings"

	//"github.com/davecgh/go-spew/spew"
	"github.com/dansteen/constellation/rkt"
	"github.com/dansteen/constellation/util"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the containers specified in the supplied config file",
	Long:  ``,
	Run:   stop,
}

func init() {
	RootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only stop when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func stop(cmd *cobra.Command, args []string) {
	fmt.Println("stop called")
	// get our running containers for this project
	runningPods, err := rkt.GetRunningPods(projectName)
	util.Check(err)

	// run through and stop them
	for name, pod := range runningPods.Pods {
		log.Println(name)
		command := strings.Split(fmt.Sprintf("rkt stop %s", pod.Name), " ")

		log.Printf("Running: %+v", command)
		stopCmd := exec.Command(command[0], command[1:]...)
		output, err := stopCmd.CombinedOutput()
		log.Printf("%s", output)
		util.Check(err)

		log.Printf("Stopped %s", name)
	}
}
