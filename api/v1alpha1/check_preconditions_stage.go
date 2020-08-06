package v1alpha1

import (
	"encoding/json"
	"github.com/fatih/structs"
)

type CheckPreconditions struct {
	Stage `json:",inline"`
	Type  string `json:"type"`
	// +optional
	Preconditions []*Precondition `json:"preconditions,omitempty"`
}

// Precondition TODO likely needs to be refined to support more than expressions
type Precondition struct {
	Context      `json:"context"`
	FailPipeline bool   `json:"failPipeline"`
	Type         string `json:"type"`
}

type Context struct {
	Expression string `json:"expression"`
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

func (cp *CheckPreconditions) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, cp)

	if err != nil {
		log.WithName("CheckPreconditions").Error(err, "error while reading CheckPreconditions")
		return err
	}

	if len(cp.Preconditions) == 0 {
		cp.Preconditions = []*Precondition{}
	}

	return nil
}

func (cp *CheckPreconditions) MarshallToMap() map[string]interface{} {
	s := structs.New(cp)
	s.TagName = "json"
	stage := s.Map()
	for key, element := range cp.Stage.MarshallToMap() {
		stage[key] = element
	}
	return stage
}
