package types

// Mount defines a volume that is mounted into a container
type Mount struct {
	Volume string
	Path   string
}
