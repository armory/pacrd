package v1alpha1

import (
	"encoding/json"

	"github.com/fatih/structs"
)

type DeleteManifest struct {
	Stage         `json:",inline"`
	Type          string `json:"type"`
	Account       string `json:"account"`
	App           string `json:"app"`
	CloudProvider string `json:"cloudProvider"`
	Location      string `json:"location"`
	// +optional
	//This should be fixed to use type DeleteManifestMode
	DeleteManifestMode `json:"mode,omitempty"`
	//This should be fixed to use type SpinnakerKind
	// +optional
	KubernetesKind `json:"kind,omitempty"`
	// +optional
	TargetName string `json:"targetName,omitempty"`
	// +optional
	LabelSelector `json:"labelSelectors,omitempty"`
	// +optional
	Options *Options `json:"options,omitempty"`
	// +optional
	Cluster string `json:"cluster,omitempty"`
	// +optional
	TargetCriteria `json:"criteria,omitempty"`

	// +optional
	Kinds []KubernetesKind `json:"kinds,omitempty"`
	//Kinds []SpinnakerKind `json:"kinds,omitempty"`
}

//Not sure where these values are in the service, need to find more but for the moment this are all possible
// +kubebuilder:validation:Enum=static;dynamic;label
type DeleteManifestMode string

const (
	// ChooseStaticTarget selector for delete manifest
	ChooseStaticTarget DeleteManifestMode = "static"
	// ChooseTargetDynamically selector for delete manifest
	ChooseTargetDynamically DeleteManifestMode = "dynamic"
	// MatchTargetLabel selector for delete manifest
	MatchTargetLabel DeleteManifestMode = "label"
)

//These values can be found in: /clouddriver/clouddriver-kubernetes-v2/src/main/java/com/netflix/spinnaker/clouddriver/kubernetes/v2/controllers/ManifestController.java
// +kubebuilder:validation:Enum=oldest;smallest;newest;largest;second_newest
type TargetCriteria string

type LabelSelector struct {
	Selector []Selector `json:"selectors,omitempty"`
}
type Options struct {
	// +optional
	Cascading bool `json:"cascading"`
	// +optional
	GracePeriodSeconds *int `json:"gracePeriodSeconds"`
}

//This value comes from: /clouddriver/clouddriver-kubernetes-v2/src/main/java/com/netflix/spinnaker/clouddriver/kubernetes/v2/security/KubernetesSelector.java
// +kubebuilder:validation:Enum=ANY;EQUALS;NOT_EQUALS;CONTAINS;NOT_CONTAINS;EXISTS;NOT_EXISTS
type SelectorsKind string

type Selector struct {
	Key           string `json:"key,omitempty"`
	SelectorsKind `json:"kind,omitempty"`
	Values        []string `json:"values,omitempty"`
}

func (dm *DeleteManifest) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, dm)

	if err != nil {
		log.WithName("DeleteManifest").Error(err, "error while reading DeleteManifest")
		return err
	}

	return nil
}

func (dm *DeleteManifest) MarshallToMap() map[string]interface{} {
	s := structs.New(dm)
	s.TagName = "json"
	stage := s.Map()
	for key, element := range dm.Stage.MarshallToMap() {
		stage[key] = element
	}

	//When we have static target the manifestname is the union of kind and targetName
	if modevalue, ok := stage["mode"]; ok && modevalue == ChooseStaticTarget {
		manifestName, err := GenerateManifestName(stage)

		if err != nil {
			log.WithName("DeleteManifest").Error(err, "error while trying to generate manifestName")
			return stage
		}

		stage["manifestName"] = manifestName
	}

	return stage
}
