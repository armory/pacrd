package v1alpha1

import (
	"encoding/json"
	"github.com/fatih/structs"
)

// StatusUrlResolution will poll a status url to determine the progress of the stage.
// +kubebuilder:validation:Enum=locationHeader;getMethod;webhookResponse
type StatusUrlResolution string

// Webhook represents a webhook stage in Spinnaker.
// NOTE: notifications currently not supported for this stage.
type Webhook struct {
	Stage         `json:",inline"`
	Type          string `json:"type"`
	Url           string `json:"url,omitempty"`
	WebhookMethod string `json:"method,omitempty"`
	// +optional
	FailOnFailedExpressions bool `json:"failOnFailedExpressions,omitempty"`
	// +optional
	ExpectedArtifacts []Artifact `json:"expectedArtifacts,omitempty"`
	// +optional
	StageEnabled `json:"stageEnabled,omitempty"`
	// +optional
	CancelEndpoint string `json:"cancelEndpoint,omitempty"`
	// +optional
	CancelMethod string `json:"cancelMethod,omitempty"`
	//+optional
	CancelPayload string `json:"cancelPayload,omitempty"`
	//+optional
	CanceledStatuses string `json:"canceledStatuses,omitempty"`
	//+optional
	CustomHeaders string `json:"customHeaders,omitempty"`
	//+optional
	Payload string `json:"payload,omitempty"`
	//+optional
	ProgressJsonPath string `json:"progressJsonPath,omitempty"`
	//+optional
	RetryStatusCodes []int `json:"retryStatusCodes,omitempty"`
	//+optional
	FailFastStatusCodes []int `json:"failFastStatusCodes,omitempty"`
	//+optional
	StatusJsonPath string `json:"statusJsonPath,omitempty"`
	//+optional
	StatusUrlResolution `json:"statusUrlResolution,omitempty"`
	//+optional
	SuccessStatuses string `json:"successStatuses,omitempty"`
	//+optional
	TerminalStatuses string `json:"terminalStatuses,omitempty"`
	//+optional
	WaitBeforeMonitor string `json:"waitBeforeMonitor,omitempty"`
	//+optional
	WaitForCompletion bool `json:"waitForCompletion,omitempty"`
	//+optional
	StatusUrlJsonPath string `json:"statusUrlJsonPath,omitempty"`
	//+optional
	SignalCancellation bool `json:"signalCancellation,omitempty"`
}

func (w *Webhook) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, w)

	if err != nil {
		log.WithName("Webhook").Error(err, "error while reading Webhook")
		return err
	}

	return nil
}

func (w *Webhook) MarshallToMap() map[string]interface{} {
	s := structs.New(w)
	s.TagName = "json"

	stage := s.Map()
	for key, element := range w.Stage.MarshallToMap() {
		stage[key] = element
	}

	err := rewriteStringValueFromMapToMapInterface("payload", stage)
	if err != nil {
		log.WithName("Webhook").Error(err, "error while reading payload")
	}
	err = rewriteStringValueFromMapToMapInterface("cancelPayload", stage)
	if err != nil {
		log.WithName("Webhook").Error(err, "error while reading cancelPayload")
	}
	err = rewriteStringValueFromMapToMapInterface("customHeaders", stage)
	if err != nil {
		log.WithName("Webhook").Error(err, "error while reading customHeaders")
	}
	return stage

}
