package engine

import (
	coreV1 "k8s.io/api/core/v1"
)

type KubeAppSpec struct {
	Image   string       `json:"image,omitempty"`
	Ports   []KubePort   `json:"ports,omitempty"`
	Volumes []KubeVolume `json:"volumes,omitempty"`
	Command []string     `json:"command,omitempty"`
	Service *KubeService `json:"service, omitempty"`
}

type KubePort struct {
	Name          string          `json:"name, omitempty"`
	HostPort      int32           `json:"hostPort,omitempty"`
	ContainerPort int32           `json:"containerPort,omitempty"`
	Protocol      coreV1.Protocol `json:"protocol,omitempty"`
}

type KubeVolume struct {
	Name         string              `json:"name,omitempty"`
	MountPath    string              `json:"mountPath,omitempty"`
	HostPath     string              `json:"hostPath, omitempty"`
	HostPathType coreV1.HostPathType `json:"hostPathType, omitempty"`
	ConfigName   string              `json:"configName, omitempty"`
	SecretName   string              `json:"SecretName, omitempty"`
}

type KubeService struct {
	Type  coreV1.ServiceType `json:"type, omitempty"`
	Ports []KubeServicePort  `json:"ports,omitempty"`
}

type KubeServicePort struct {
	Name       string          `json:"name,omitempty"`
	Port       int32           `json:"port,omitempty"`
	TargetPort int32           `json:"targetPort, omitempty"`
	NodePort   int32           `json:"targetPort, omitempty"`
	Protocol   coreV1.Protocol `json:"protocol,omitempty"`
}

type NativeAppSpec struct {
	Rc      *NativeResource `json:"rc,omitempty"`
	Command []string        `json:"command,omitempty"`
}

type NativeResource struct {
	Type     string `json:"type,omitempty"`
	FileName string `json:"fileName,omitempty"`
	Url      string `json:"url,omitempty"`
	Md5      string `json:"md5,omitempty"`
}
