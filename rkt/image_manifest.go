package rkt

// ImageManifest holds information about an image.  Note: we only grab what we need at this point (or what is easy)
type ImageManifest struct {
	ACKind    string   `json:"acKind"`
	ACVersion string   `json:"acVersion"`
	Name      string   `json:"name"`
	App       ImageApp `json:"app"`
}

// ImageApp holds information from the image-manifest about an app.  Note: we only grab what we need at this point (or what is easy)
type ImageApp struct {
	Ports []ImageAppPort `json:"ports"`
}

// ImageAppPort is the representation of port information found in rkt image manifests
type ImageAppPort struct {
	Name            string `json:"name"`
	Protocol        string `json:"protocol"`
	Port            int    `json:"port"`
	count           int    `json:"count"`
	socketActivated bool   `json:"socketActivated"`
}
