package v1alpha1

import (
	"encoding/json"

	"github.com/fatih/structs"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:generate=false
type SpinnakerTrigger interface {
	NewTriggerFromBytes([]byte) error
	MarshallToMap() map[string]interface{}
}

type Trigger struct {
	Type       string               `json:"type" yaml:"type" protobuf:"bytes,2,name=type"`
	Properties runtime.RawExtension `json:"properties,omitempty"`
}

// +kubebuilder:object:generate=false
type DockerTrigger struct {
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
	Organization string `json:"organization"`
	Registry     string `json:"registry"`
	Repository   string `json:"repository"`
	Account      string `json:"account"`
}

func (dt *DockerTrigger) NewTriggerFromBytes(bs []byte) error {
	return json.Unmarshal(bs, dt)
}

func (dt *DockerTrigger) MarshallToMap() map[string]interface{} {
	s := structs.New(dt)
	s.TagName = "json"
	return s.Map()
}

// +kubebuilder:object:generate=false
type UnknownTrigger struct {
	Type       string                 `json:"type,omitempty"`
	Properties map[string]interface{} `json:"properties"`
}

func (ct *UnknownTrigger) NewTriggerFromBytes(bs []byte) error {
	m := make(map[string]interface{})

	err := json.Unmarshal(bs, &m)
	ct.Properties = m

	return err
}

func (ct *UnknownTrigger) MarshallToMap() map[string]interface{} {
	ct.Properties["type"] = ct.Type // Persist the type we got at ingest time.
	return ct.Properties
}

func (t Trigger) ToTrigger() (SpinnakerTrigger, error) {

	var s SpinnakerTrigger

	switch t.Type {
	case "docker":
		s = &DockerTrigger{Type: t.Type}
	default:
		s = &UnknownTrigger{Type: t.Type}
	}

	err := s.NewTriggerFromBytes(t.Properties.Raw)
	return s, err
}
