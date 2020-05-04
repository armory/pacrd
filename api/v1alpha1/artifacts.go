package v1alpha1

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/structs"
	"k8s.io/apimachinery/pkg/runtime"
)

// Artifact is an object that references an external resource. It could be a
// Docker container, file in source control, AMI, or binary blob in S3, etc.
type Artifact struct {
	// ID is a unique identifier for this artifact. IDs must only be unique for
	// the pipeline they are declared in.
	ID string `json:"id"`
	// DisplayName tells Spinnaker how to render this artifact in the UI.
	DisplayName string `json:"displayName"`
	// Attempt to match against an artifact in the prior pipeline execution's context.
	//
	// See the [reference](https://www.spinnaker.io/reference/artifacts/in-pipelines)
	// for more information.
	// +optional
	UsePriorArtifact bool `json:"usePriorArtifact,omitempty"`
	// If true, requires DefaultArtifact to be defined with a fallback artifact to use.
	// +optional
	UseDefaultArtifact bool `json:"useDefaultArtifact,omitempty"`
	// If your artifact either wasn't supplied from a trigger, or it wasn't found
	// in a prior execution, the artifact specified here will end up in your
	// pipeline's execution context.
	// +optional
	DefaultArtifact *MatchArtifact `json:"defaultArtifact,omitempty"`
	// This specifies which fields in your incoming artifact to match against.
	// Every field that you supply will be used to match against all incoming
	// artifacts. If all specified fields match, the incoming artifact is bound
	// to your pipeline context.
	//
	// See the [reference](https://www.spinnaker.io/reference/artifacts/in-pipelines/#expected-artifacts)
	// for more information.
	// +optional
	MatchArtifact `json:"matchArtifact,omitempty"`
}

func (a Artifact) MarshallToMap() (map[string]interface{}, error) {
	s := structs.New(a)
	s.TagName = "json"
	artifactMap := s.Map() // Take a first past at serialization.

	matchArtifact, err := a.MatchArtifact.ToMatchArtifact()
	if err != nil {
		return map[string]interface{}{}, err // TODO wrap error
	}
	artifactMap["matchArtifact"] = matchArtifact.MarshallToMap()

	if a.DefaultArtifact != nil {
		defaultArtifact, err := a.DefaultArtifact.ToMatchArtifact()
		if err != nil {
			return map[string]interface{}{}, err // TODO wrap error
		}
		artifactMap["defaultArtifact"] = defaultArtifact.MarshallToMap()
	}

	return artifactMap, nil
}

type ArtifactReference struct {
	// +optional
	ID string `json:"id,omitempty"`
	// +optional
	DisplayName string `json:"displayName,omitempty"`
}

type ErrNameUndefined struct {
	ArtifactName string
}

func (e *ErrNameUndefined) Error() string {
	return fmt.Sprintf("artifact %q must have a non-empty name", e.ArtifactName)
}

// SpinnakerMatchArtifact represents TODO
// +kubebuilder:object:generate=false
type SpinnakerMatchArtifact interface {
	NewFromBytes([]byte) error
	MarshallToMap() map[string]interface{}
}

// +kubebuilder:object:generate=false
type UnknownArtifact struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

func (ua *UnknownArtifact) NewFromBytes(data []byte) error {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	ua.Properties = m

	return err
}

func (ua *UnknownArtifact) MarshallToMap() map[string]interface{} {
	ua.Properties["type"] = ua.Type // Persist the type we got at ingest time.

	return ua.Properties
}

type CustomArtifact struct {
	Type            string `json:"type"`
	Name            string `json:"name"`
	ID              string `json:"id,omitempty"`
	Location        string `json:"location,omitempty"`
	Reference       string `json:"reference,omitempty"`
	Version         string `json:"version,omitempty"`
	ArtifactAccount string `json:"artifactAccount,omitempty"`
}

func (ca *CustomArtifact) NewFromBytes(data []byte) error {
	err := json.Unmarshal(data, ca)

	if err != nil {
		return err
	}

	if ca.Name == "" {
		return &ErrNameUndefined{"CustomArtifact"}
	}

	return nil
}

// EmbeddedArtifact represents a base64 encoded artifact.
type EmbeddedArtifact struct {
	Type string `json:"type"`
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

func (ea *EmbeddedArtifact) NewFromBytes(data []byte) error {
	err := json.Unmarshal(data, ea)

	if err != nil {
		return err
	}

	if ea.Name == "" {
		return &ErrNameUndefined{"EmbeddedArtifact"}
	}

	return nil
}

func (ea *EmbeddedArtifact) MarshallToMap() map[string]interface{} {
	s := structs.New(ea)
	s.TagName = "json"
	return s.Map()
}

// DockerArtifact represents a container in the target Docker registry.
// +kubebuilder:object:generate=false
type DockerArtifact struct {
	// ArtifactAccount represents the desired container registry to pull images from.
	ArtifactAccount string `json:"artifactAccount"` // TODO could be validated further
	// Name is the fully qualified Docker image name in the configured registry.
	Name string `json:"name"`
	// ID represents a pipeline-wide unique identifier.
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

func (da *DockerArtifact) MarshallToMap() map[string]interface{} {
	s := structs.New(da)
	s.TagName = "json"
	return s.Map()
}

func (da *DockerArtifact) NewFromBytes(data []byte) error {
	err := json.Unmarshal(data, da)

	if err != nil {
		return err
	}

	if da.ID == "" {
		// TODO if id is not provided, compute one via uuid
	}

	if da.Name == "" {
		return &ErrNameUndefined{"DockerArtifact"}
	}

	if da.ArtifactAccount == "" {
		return fmt.Errorf("artifactAccount must be defined for this artifact")
	}

	return nil
}

type MatchArtifact struct {
	Type string `json:"type" yaml:"type" protobuf:"bytes,2,name=type"`
	// +optional
	Properties runtime.RawExtension `json:"properties,omitempty"`
}

func (ma MatchArtifact) ToMatchArtifact() (SpinnakerMatchArtifact, error) {
	var a SpinnakerMatchArtifact

	switch ma.Type {
	case "docker/image":
		a = &DockerArtifact{Type: ma.Type}
	case "embedded/base64":
		a = &EmbeddedArtifact{Type: ma.Type}
	default:
		a = &UnknownArtifact{Type: ma.Type}
	}

	err := a.NewFromBytes(ma.Properties.Raw)
	return a, err
}

// GetArtifactID returns the ID of the given artifact if defined in this pipeline.
func GetArtifactID(artifacts []Artifact, r ArtifactReference) (string, error) {
	for _, a := range artifacts {
		if r.ID != "" && r.ID == a.ID {
			return a.ID, nil
		}

		if r.DisplayName != "" && r.DisplayName == a.DisplayName {
			return a.ID, nil
		}
	}

	return "", fmt.Errorf("artifact with id %q and name %q could not be found for this pipeline", r.ID, r.DisplayName)
}
