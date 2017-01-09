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
	"os"

	"github.com/dansteen/constellation/config"
	"github.com/dansteen/constellation/types"
	"github.com/dansteen/constellation/util"
	//"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
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
	configPath := fmt.Sprintf("/tmp/constellation-%d", os.Getpid())
	fmt.Println("run called")
	// generate our network file
	network, err := types.NewNetworkConfig(projectName)
	util.Check(err)
	// convert our config to json
	networkJSON, err := json.MarshalIndent(network, "", "  ")
	// make our config folder
	util.Check(os.MkdirAll(fmt.Sprintf("%s/net.d", configPath), 0755))
	// write our network files
	util.Check(ioutil.WriteFile(fmt.Sprintf("%s/net.d/%s.conf", configPath, projectName), networkJSON, 0644))
	// process our configs
	configData := config.ProcessFile(podFile, includeDirs)

	// initialize the containers
	for _, container := range configData.Containers {
		util.Check(container.Init(configData.Containers))
	}

	// make sure to create our log volumes
	for _, volume := range configData.Volumes {
		util.Check(volume.CreateDir())
	}

	// determin the order we need to execute in to satisfy dependencies
	order, err := configData.DependencyOrder()
	util.Check(err)
	//spew.Dump(configData)

	for _, containerName := range order {
		// grab our container
		container := configData.Containers[containerName]
		// run our container
		err := container.Run(configPath, projectName, configData.Volumes)
		util.Check(err)

	}

	fmt.Println(order)
}
