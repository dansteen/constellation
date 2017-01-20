package types

import (
	"fmt"
	"log"
	"os"
)

// Volume defines the location that mounts will mount from on the host machine
type Volume struct {
	Name string      `json:"name"`
	Kind string      `json:"kind"`
	Path string      `json:"path"`
	UID  int         `json:"uid"`
	GID  int         `json:"gid"`
	Mode os.FileMode `json:"mode"`
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
	log.Printf("Creating volume %s at %s.  ", volume.Name, volume.Path)
	if volume.Kind == "host" {
		// create the folder
		err := os.MkdirAll(volume.Path, volume.Mode)
		// once we've done that update ownership and mode
		if err != nil {
			return err
		}
		log.Printf("Changing Ownership to %d:%d\n", volume.UID, volume.GID)
		err = os.Chown(volume.Path, volume.UID, volume.GID)
		if err != nil {
			return err
		}
		log.Printf("Changing Mode.\n")
		err = os.Chmod(volume.Path, volume.Mode)
		if err != nil {
			return err
		}
	}
	return nil
}
