package container

import "fmt"

// Mount defines a volume that is mounted into a container
type Mount struct {
	Volume string
	Path   string
}

func (mount *Mount) GenerateCommandLine() []string {
	mountArray := make([]string, 2)
	mountArray[0] = "--mount"
	mountArray[1] = fmt.Sprintf("volume=%s,target=%s", mount.Volume, mount.Path)
	return mountArray
}
