package types

import (
	"fmt"
	"os"
)

// Volume defines the location that mounts will mount from on the host machine
type Volume struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Path string `json:"path"`
}

// GenerateCommandLine generates the command line flags for this mount
func (volume *Volume) GenerateCommandLine() []string {
	volumeArray := make([]string, 2)
	volumeArray[0] = "--volume"
	volumeArray[1] = fmt.Sprintf("%s,kind=%s,source=%s", volume.Name, volume.Kind, volume.Path)
	return volumeArray
}

// CreateDir creates the directory pointed to in this volume if it does not exist
// this is only done if the volume.Kind is set to "host".
func (volume *Volume) CreateDir() error {
	if volume.Kind == "host" {
		return os.MkdirAll(volume.Path, 0755)
	}
	return nil
}
