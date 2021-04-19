package v1alpha1

import (
	"encoding/json"

	"github.com/fatih/structs"
)

// BakeManifest represents a bake manifest stage in Spinnaker.
// NOTE: I suspect this only supports `helm2` style deployments right now.
// NOTE: notifications currently not supported for this stage.
type BakeManifest struct {
	Stage `json:",inline"`

	// Name is the name given to this stage.
	Type string `json:"type"`

	// +optional
	Namespace string `json:"namespace,omitempty"`
	// +optional
	EvaluateOverrideExpressions bool `json:"evaluateOverrideExpressions,omitempty"`
	// +optional
	ExpectedArtifacts []Artifact `json:"expectedArtifacts,omitempty"`
	// +optional
	InputArtifacts []*ArtifactReference `json:"inputArtifacts,omitempty"`
	// InputArtifact is used by the Kustomize variant of BakeManifest to pull in a single artifact.
	// +optional
	InputArtifact ArtifactReference `json:"inputArtifact,omitempty"`
	// +optional
	OutputName string `json:"outputName,omitempty"`
	// +optional
	Overrides map[string]string `json:"overrides,omitempty"`
	// +optional
	RawOverrides     bool   `json:"rawOverrides,omitempty"`
	TemplateRenderer string `json:"templateRenderer,omitempty"`
	// KustomizeFilePath is the relative path to the kustomize.yaml file in the given artifact.
	// +optional
	KustomizeFilePath string `json:"kustomizeFilePath,omitempty"`
}

func (bk *BakeManifest) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, bk)

	if err != nil {
		log.WithName("BakeManifest").Error(err, "error while reading BakeManifest")
		return err
	}

	if len(bk.InputArtifacts) == 0 {
		bk.InputArtifacts = []*ArtifactReference{}
	}

	if len(bk.Overrides) == 0 {
		bk.Overrides = make(map[string]string)
	}

	return nil
}

func (bk *BakeManifest) MarshallToMap() map[string]interface{} {
	s := structs.New(bk)
	s.TagName = "json"
	stage := s.Map()
	for key, element := range bk.Stage.MarshallToMap() {
		stage[key] = element
	}
	return stage
}
