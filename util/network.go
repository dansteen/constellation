package util

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strings"
)

// GetDefaultIP will return the ip address associated with the default interface.  Unfortunately, there is no
// current way to get this information using standard system calls.  so we read from proc, but we don't ever throw
// an error.  We just return an empty string.
func GetDefaultIP() string {
	// the format of the routes file
	const (
		Iface = iota
		Destination
		Gateway
		Flags
		RefCnt
		Use
		Metric
		Mask
		MTU
		Window
		IRTT
	)
	// read from the system routes file
	routeFilePath := "/proc/net/route"

	routeFile, err := os.Open(routeFilePath)
	if err != nil {
		return ""
	}

	routesReader := bufio.NewReader(routeFile)
	routesCsv := csv.NewReader(routesReader)
	routesCsv.Comma = '\t'
	routesCsv.Comment = '#'
	routesCsv.TrimLeadingSpace = true

	routes, err := routesCsv.ReadAll()
	if err != nil {
		return fmt.Sprintf("%s", err)
	}

	// once we have our routes we grab the default and get the IP of that interface
	for _, route := range routes {
		if route[Destination] == "00000000" {
			iface, err := net.InterfaceByName(route[Iface])
			if err != nil {
				return ""
			}
			// we return the first address
			addresses, err := iface.Addrs()
			if err != nil {
				return ""
			}
			for _, address := range addresses {
				// strip off the trailing CIDR notation
				return strings.Split(address.String(), "/")[0]
			}
			return ""
		}
	}
	return ""
}
