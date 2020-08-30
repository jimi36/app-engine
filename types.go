package engine

import (
	"strings"
)

type EngineType string

const (
	// Kube is use kubernetes
	Kube EngineType = "kube"
	// Native is native process
	Native EngineType = "native"
)

type Application struct {
	ApplicationTag `json:",inline"`

	Labels map[string]string `json:"labels,omitempty"`

	Env map[string]string `json:"env,omitempty"`

	// engine type
	Type EngineType `json:"type, omitempty"`

	KubeSpec   *KubeAppSpec   `json:"kubeSpec,omitempty"`
	NativeSpec *NativeAppSpec `json:"nativeSpec,omitempty"`
}

type ApplicationList []*Application

func (this ApplicationList) Len() int {
	return len(this)
}
func (this ApplicationList) Less(i, j int) bool {
	return strings.Compare(this[i].Name, this[j].Name) < 0
}
func (this ApplicationList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type ApplicationTag struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

func (tag *ApplicationTag) Tag() string {
	return strings.Join([]string{tag.Name, tag.Version}, "-")
}

type ApplicationRuntime struct {
	ApplicationTag `json:",inline"`

	ToStart   bool   `json:"toStart,omitempty"`
	IsStarted bool   `json:"isStarted,omitempty"`
	Err       string `json:"err,omitempty"`

	// For native
	Pid int `json:"pid,omitempty"`
}

type ApplicationState struct {
	Name      string          `json:"name,omitempty"`
	Version   string          `json:"version,omitempty"`
	ToStart   bool            `json:"toStart,omitempty"`
	IsStarted bool            `json:"isStarted,omitempty"`
	Err       string          `json:"err,omitempty"`
	Instances []InstanceState `json:"instances,omitempty"`
}

type InstanceState struct {
	Name    string `json:"name,omitempty"`
	Running bool   `json:"running,omitempty"`
	Cpu     int64  `json:"cpu,omitempty"`
	Mem     int64  `json:"mem,omitempty"`
}

type Config struct {
	Name   string            `json:"name,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
	Data   map[string]string `json:"data,omitempty"`
}

type ListApplicationOption struct {
	Size    int    `json:"size"`
	LastPos string `json:"lastPos"`
}
