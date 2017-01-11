package types

import (
	"errors"
	"fmt"
	"strings"
)

// Hosts defines an entry in the hosts file
type HostsEntry struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}

// GenerateCommandLine generates the command line flags for this mount
func (hosts *HostsEntry) GenerateCommandLine() []string {
	hostsArray := make([]string, 2)
	hostsArray[0] = "--hosts-entry"
	hostsArray[1] = fmt.Sprintf("%s=%s", hosts.IP, hosts.Name)
	return hostsArray
}

// HostsEntryFromString will generate a hostsEntry from a string in IP=NAME format
func HostsEntryFromString(entry string) (HostsEntry, error) {
	entryArray := strings.SplitN(entry, "=", 2)
	if len(entryArray) != 2 {
		return HostsEntry{}, errors.New(fmt.Sprintf("HostsEntries need to be in IP=Name format.  Got %s", entry))
	}
	return HostsEntry{
		IP:   entryArray[0],
		Name: entryArray[1],
	}, nil
}
