package types

// Mount defines a volume that is mounted into a container
type Mount struct {
	volume string
	path   string
}

// returns a new Mount struct populated with the supplied values
func NewMount(volume string, path string) Mount {
	return Mount{
		volume: volume,
		path:   path,
	}
}
