package v1alpha1

import (
	"encoding/json"

	"github.com/fatih/structs"
)

// FindArtifactsFromResource represents the stage of the same name in Spinnaker.
type FindArtifactsFromResource struct {
	Stage         `json:",inline"`
	Type          string `json:"type"`
	Account       string `json:"account"`
	App           string `json:"app,omitempty"`
	CloudProvider string `json:"cloudProvider"`
	Location      string `json:"location"`
	Mode          string `json:"mode"` // FIXME enum static/dynamic
	ManifestName  string `json:"manifestName"`
}

func (fafr *FindArtifactsFromResource) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, fafr)

	if err != nil {
		log.WithName("FindArtifactsFromResource").Error(err, "error while reading FindArtifactsFromResource")
		return err
	}

	return nil
}

func (fafr *FindArtifactsFromResource) MarshallToMap() map[string]interface{} {
	s := structs.New(fafr)
	s.TagName = "json"
	stage := s.Map()
	for key, element := range fafr.Stage.MarshallToMap() {
		stage[key] = element
	}
	return stage
}
