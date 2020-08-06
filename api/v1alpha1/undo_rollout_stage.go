package v1alpha1

import (
	"encoding/json"
	"github.com/fatih/structs"
)

// UndoRolloutManifestMode is the means for undoing a manifest rollout.
// +kubebuilder:validation:Enum=static
type UndoRolloutManifestMode string

const (
	// UndoRolloutManifestStaticMode .
	UndoRolloutManifestStaticMode UndoRolloutManifestMode = "static"
)

// UndoRolloutManifest is a stage that rolls back a manifest.
type UndoRolloutManifest struct {
	Stage            `json:",inline"`
	Type             string `json:"type"`
	Account          string `json:"account"`
	CloudProvider    string `json:"cloudProvider"`
	Location         string `json:"location"`
	NumRevisionsBack int    `json:"numRevisionsBack"`
	// +optional
	Mode UndoRolloutManifestMode `json:"mode,omitempty"`
	// +optional
	TargetName string `json:"targetName,omitempty"`
	// +optional
	Kind KubernetesKind `json:"kind,omitempty"`
}

func (urm *UndoRolloutManifest) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, urm)

	if err != nil {
		log.WithName("UndoRolloutManifest").Error(err, "error while reading UndoRolloutManifest")
		return err
	}

	return nil
}

func (urm *UndoRolloutManifest) MarshallToMap() map[string]interface{} {
	s := structs.New(urm)
	s.TagName = "json"
	stage := s.Map()
	for key, element := range urm.Stage.MarshallToMap() {
		stage[key] = element
	}

	//When we have static target the manifestname is the union of kind and targetName
	if modevalue, ok := stage["mode"]; ok && modevalue == UndoRolloutManifestStaticMode {
		manifestName, err := GenerateManifestName(stage)

		if err != nil {
			log.WithName("UndoRolloutManifest").Error(err, "error while trying to generate manifestName")
			return stage
		}

		stage["manifestName"] = manifestName
	}

	return stage
}
