package v1alpha1

import "github.com/fatih/structs"

type Stage struct {
	// Name is the name given to this stage.
	Name string `json:"name"`
	// RefID is the position in the pipeline graph that this stage should live. Usually monotonically increasing for a pipeline.
	RefID string `json:"refId"`
	// RequisiteStageRefIds is a list of RefIDs that are required before this stage can run.
	// +optional
	RequisiteStageRefIds []string `json:"requisiteStageRefIds,omitempty"`
	// StageEnabled represents whether this stage is active in a pipeline graph.
	// +optional
	StageEnabled *StageEnabled `json:"stageEnabled,omitempty"`
	// Comments provide additional context for this stage in the Spinnaker UI.
	// +optional
	Comments string `json:"comments,omitempty"`
	// RestrictExecutionDuringTimeWindow provides the ability to restrict the hours during which this stage can run.
	// +optional
	RestrictExecutionDuringTimeWindow bool `json:"restrictExecutionDuringTimeWindow,omitempty"`

	// RestrictedExecutionWindow provides the ability to restrict the hours during which this stage can run.
	// +optional
	RestrictedExecutionWindow RestrictedExecutionWindow `json:"restrictedExecutionWindow,omitempty"`
	// SkipWindowText is the text to display when this stage is skipped.
	// +optional
	SkipWindowText string `json:"skipWindowText,omitempty"`

	// +optional
	FailOnFailedExpressions bool `json:"failOnFailedExpressions,omitempty"`
	// +optional
	FailPipeline *bool `json:"failPipeline,omitempty"`
	// +optional
	ContinuePipeline *bool `json:"continuePipeline,omitempty"`
	// +optional
	CompleteOtherBranchesThenFail *bool `json:"completeOtherBranchesThenFail,omitempty"`
}

// StageEnabled represents whether this stage is active in a pipeline graph.
type StageEnabled struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

// RestrictedExecutionWindow TODO description
type RestrictedExecutionWindow struct {
	Days      []int `json:"days,omitempty"` // TODO candidate for further validation
	Jitter    `json:"jitter,omitempty"`
	WhiteList []WhiteListWindow `json:"whitelist,omitempty"`
}

// WhiteListWindow TODO description
type WhiteListWindow struct {
	EndHour   int `json:"endHour,omitempty"`
	EndMin    int `json:"endMin,omitempty"`
	StartHour int `json:"startHour,omitempty"`
	StartMin  int `json:"startMin,omitempty"`
}

// Jitter TODO description
type Jitter struct {
	Enabled    bool `json:"enabled,omitempty"`
	MaxDelay   int  `json:"maxDelay,omitempty"`
	MinDelay   int  `json:"minDelay,omitempty"`
	SkipManual bool `json:"skipManual,omitempty"`
}

func (s *Stage) MarshallToMap() map[string]interface{} {

	if len(s.RequisiteStageRefIds) == 0 {
		s.RequisiteStageRefIds = []string{}
	}

	m := structs.New(s)
	m.TagName = "json"
	return m.Map()

}
