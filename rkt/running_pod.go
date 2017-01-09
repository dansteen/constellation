package rkt

import (
	"encoding/json"
	"time"
)

// Pod holds information about containers
type Pod struct {
	Name      string    `json:"name"`
	State     string    `json:"state"`
	Networks  []Network `json:"networks"`
	AppNames  []string  `json:"app_names"`
	StartedAt time.Time `json:"started_at"`
}

// UnmarshalJSON
func (pod *Pod) UnmarshalJSON(b []byte) error {
	// create a struct to unmarshal into
	type stringPod struct {
		Name     string    `json:"name"`
		State    string    `json:"state"`
		Networks []Network `json:"networks"`
		AppNames []string  `json:"app_names"`

		StartedAt int64 `json:"started_at"`
	}
	var tmpPod stringPod
	// unmarshal our items into the container
	err := json.Unmarshal(b, &tmpPod)
	if err != nil {
		return err
	}

	// set our pod values
	pod.Name = tmpPod.Name
	pod.State = tmpPod.State
	pod.Networks = tmpPod.Networks
	pod.AppNames = tmpPod.AppNames
	pod.StartedAt = time.Unix(tmpPod.StartedAt, 0)
	return nil
}

// Network holds network information for a running container
type Network struct {
	NetName    string `json:"netName"`
	NetConf    string `json:"netConf"`
	PluginPath string `json:"pluginPath"`
	IfName     string `json:"ifName"`
	IP         string `json:"ip"`
	Args       string `json:"args"`
	Mask       string `json:"mask"`
}
