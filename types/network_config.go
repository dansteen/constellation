package types

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

// NetworkConfig holds the config for a container network
type NetworkConfig struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Bridge    string `json:"bridge"`
	IsGateway bool   `json:"isGateway"`
	IPMasq    bool   `json:"ipMasq"`
	IPAM      IPAM   `json:"ipam"`
}

// IPAM holds ipam config for a NetworkConfig
type IPAM struct {
	Type   string              `json:"type"`
	Subnet string              `json:"subnet"`
	Routes []map[string]string `json:"routes"`
}

// NewNetworkConfig will generate a new network config object for the provided project name
func NewNetworkConfig(projectName string) (NetworkConfig, error) {
	// first we find an available subnet in the range 172.0.0.0/8
	// we just randomize and check
	// first generate random values for our middle octets
	num := time.Now().UnixNano()
	rand.Seed(num)
	// get our interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return NetworkConfig{}, err
	}
	// run through until we find an empty subnet
	subnet := ""
	for good := false; !good; {
		// a value to see if we succeed
		subnet = fmt.Sprintf("172.16.%d.0/24", rand.Intn(255))
		for _, net := range interfaces {
			addresses, err := net.Addrs()
			if err != nil {
				return NetworkConfig{}, err
			}
			for _, addr := range addresses {
				if strings.Split(addr.String(), "/")[0] != strings.Split(subnet, "/")[0] {
					good = true
					break
				}
			}
		}
	}

	routes := make(map[string]string)
	routes["dst"] = "0.0.0.0/0"
	routeArray := make([]map[string]string, 1)
	routeArray[0] = routes

	// the config is mostly the same for each project
	config := NetworkConfig{
		Name:      fmt.Sprintf("br-%s", projectName),
		Type:      "bridge",
		Bridge:    projectName,
		IsGateway: true,
		IPMasq:    true,
		IPAM: IPAM{
			Type:   "host-local",
			Subnet: subnet,
			Routes: routeArray,
		},
	}
	return config, nil
}
