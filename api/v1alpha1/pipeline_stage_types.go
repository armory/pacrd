package v1alpha1

import (
	"encoding/json"
	"github.com/fatih/structs"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("stages")

// SpinnakerMatchStage represents TODO
// +kubebuilder:object:generate=false
type SpinnakerMatchStage interface {
	NewStageFromBytes([]byte) error
	MarshallToMap() map[string]interface{}
}

type MatchStage struct {
	// +optional
	Type string `json:"type" yaml:"type" protobuf:"bytes,2,name=type"`
	// +optional
	Properties runtime.RawExtension `json:"properties,omitempty"`
}

func (su *MatchStage) MarshallToMap() map[string]interface{} {
	s := structs.New(su)
	s.TagName = "json"
	return s.Map()
}

func (su *MatchStage) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, su)

	if err != nil {
		log.WithName("MatchStage").Error(err, "error while reading MatchStage")
		return err
	}

	return nil
}

func (su *MatchStage) GetStage() interface{} {
	var s SpinnakerMatchStage
	switch su.Type {
	case "bakeManifest":
		s = &BakeManifest{Type: su.Type}
	case "findArtifactsFromResource":
		s = &FindArtifactsFromResource{Type: su.Type}
	case "manualJudgment":
		s = &ManualJudgment{Type: su.Type}
	case "deleteManifest":
		s = &DeleteManifest{Type: su.Type}
	case "undoRolloutManifest":
		s = &UndoRolloutManifest{Type: su.Type}
	case "checkPreconditions":
		s = &CheckPreconditions{Type: su.Type}
	case "deployManifest":
		s = &DeployManifest{Type: su.Type}
	case "webhook":
		s = &Webhook{Type: su.Type}
	default:
		s = &UnknownStage{Type: su.Type}
	}
	return s
}

// ToSpinnakerStage TODO description
func (su MatchStage) ToSpinnakerStage() (map[string]interface{}, error) {

	s := su.GetStage().(SpinnakerMatchStage)
	err := s.NewStageFromBytes(su.Properties.Raw)

	stage := s.MarshallToMap()
	stage["type"] = su.Type
	delete(stage, "Stage")
	return stage, err
}
