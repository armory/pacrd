package v1alpha1

import (
	"encoding/json"

	"github.com/fatih/structs"
)

// ManualJudgment TODO description
type ManualJudgment struct {
	Stage          `json:",inline"`
	Type           string `json:"type"`
	Instructions   string `json:"instructions,omitempty"`
	StageTimeoutMs int    `json:"stageTimeoutMs,omitempty"`
	//+optional
	SendNotifications bool `json:"sendNotifications,omitempty"`
	//+optional
	Notifications []*ManualJudgmentNotification `json:"notifications,omitempty"`
	//+optional
	JudgmentInputs []JudgmentInput `json:"judgmentInputs,omitempty"`
}

// JudgmentInput TODO description
type JudgmentInput struct {
	Value string `json:"value"`
}

// ManualJudgmentNotification TODO description
type ManualJudgmentNotification struct {
	Type    string           `json:"type"`
	Address string           `json:"address,omitempty"`
	Level   string           `json:"level,omitempty"`
	Message *JudgmentMessage `json:"message,omitempty"`
	When    *[]JudgmentState `json:"when,omitempty"`
}

// JudgmentMessage TODO description
type JudgmentMessage struct {
	ManualJudgmentContinue *JudgmentMessageValue `json:"manualJudgmentContinue,omitempty"`
	ManualJudgmentStop     *JudgmentMessageValue `json:"manualJudgmentStop,omitempty"`
}

// JudgmentMessageValue TODO description
type JudgmentMessageValue struct {
	Text string `json:"text"`
}

// JudgmentState TODO description
// +kubebuilder:validation:Enum=ManualJudgmentState;ManualJudgmentContinueState;ManualJudgmentStopState
type JudgmentState string

const (
	// ManualJudgmentState for notifications
	ManualJudgmentState JudgmentState = "manualJudgment"
	// ManualJudgmentContinueState for notifications
	ManualJudgmentContinueState JudgmentState = "manualJudgmentContinue"
	// ManualJudgmentStopState for notifications
	ManualJudgmentStopState JudgmentState = "manualJudgmentStop"
)

func (mj *ManualJudgment) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, mj)

	if err != nil {
		log.WithName("ManualJudgment").Error(err, "error while reading ManualJudgment")
		return err
	}

	if len(mj.Notifications) == 0 {
		mj.Notifications = []*ManualJudgmentNotification{}
	}

	if len(mj.JudgmentInputs) == 0 {
		mj.JudgmentInputs = []JudgmentInput{}
	}

	return nil
}

func (mj *ManualJudgment) MarshallToMap() map[string]interface{} {
	s := structs.New(mj)
	s.TagName = "json"
	stage := s.Map()
	for key, element := range mj.Stage.MarshallToMap() {
		stage[key] = element
	}
	return stage
}
