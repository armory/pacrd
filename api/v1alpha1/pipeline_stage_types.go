package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"
)

// StageUnionType is an alias for the type name of the pipeline's stage.
// +kubebuilder:validation:Enum=BakeManifest;FindArtifactsFromResource;ManualJudgment;DeleteManifest;CheckPreconditions;DeployManifest
type StageUnionType string

// StageUnion is a union type that encompasses strongly typed stage defnitions.
// FIXME: in general: notifications, execution options, produces artifacts, and comments should be lifted to this type
// FIXME: if requisiteStageRefIds is not validated, pipeline will no longer render correctly in UI
type StageUnion struct {
	Type StageUnionType `json:"type"`

	// Name is the name given to this stage.
	Name string `json:"name"`
	// RefID is the position in the pipeline graph that this stage should live. Usually monotonically increasing for a pipeline.
	RefID string `json:"refId"`
	// RequisiteStageRefIds A list of RefIDs that are required before this stage can run.
	// +optional
	RequisiteStageRefIds []string `json:"requisiteStageRefIds,omitempty"`

	// +optional
	StageEnabled *StageEnabled `json:"stageEnabled,omitempty"`
	//BakeManifest renders a Kubernetes manifest to be applied to a target cluster at a later stage. The manifests can be rendered using HELM2 or Kustomize.
	// +optional
	BakeManifest `json:"bakeManifest,omitempty"`
	// +optional
	FindArtifactsFromResource `json:"findArtifactsFromResource,omitempty"`
	//ManualJudgment stage pauses pipeline execution until there is approval from a human through the UI or API call that allows the execution to proceed.
	// +optional
	ManualJudgment `json:"manualJudgment,omitempty"`
	//DeleteManifest removes a manifest or a group of manifests from a target Spinnaker cluster based on names, deployment version or labels.
	// +optional
	DeleteManifest `json:"deleteManifest,omitempty"`
	// CheckPreconditions allows you to test values from the pipeline's context to determine wether to proceed, pause, or terminate the pipeline execution
	// +optional
	CheckPreconditions `json:"checkPreconditions,omitempty"`
	// DeployManifest deploys a Kubernetes manifest to a target Kubernetes cluster. Spinnaker will periodically check the status of the manifest to make sure the manifest converges on the target cluster until it reaches a timeout
	// +optional
	DeployManifest `json:"deployManifest,omitempty"`
}

// DeployManifest TODO
// FIXME: trafficManagement, relationships
type DeployManifest struct {
	Account                       string `json:"account"`
	CloudProvider                 string `json:"cloudProvider"`
	CompleteOtherBranchesThenFail bool   `json:"completeOtherBranchesThenFail"`
	ContinuePipeline              bool   `json:"continuePipeline"`
	FailPipeline                  bool   `json:"failPipeline"`
	// +optional
	ManifestArtifactAccount string `json:"manifestArtifactAccount,omitempty"`
	// +optional
	ManifestArtifactID string `json:"manifestArtifactId,omitempty"`
	// +optional
	Manifests []string         `json:"manifests,omitempty"` // FIXME
	Moniker   `json:"moniker"` // FIXME: should be calculated
	// +optional
	SkipExpressionEvaluation bool `json:"skipExpressionEvaluation,omitempty"`
	// +optional
	Source Source `json:"source,omitempty"`
}

// Source represents the kind of DeployManifest stage is defined.
// +kubebuilder:validation:Enum=text;artifact
type Source string

const (
	// TextManifest represents inline manifests in the DeployInlineManifest stage.
	TextManifest Source = "text"
	// ArtifactManifest represents manifests that live outside the stage.
	ArtifactManifest Source = "artifact"
)

// Moniker TODO
type Moniker struct {
	App string `json:"app"`
}

type CheckPreconditions struct {
	// +optional
	Preconditions *[]Precondition `json:"preconditions,omitempty"`
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

type DeleteManifest struct {
	Account       string `json:"account"`
	App           string `json:"app"`
	CloudProvider string `json:"cloudProvider"`
	Location      string `json:"location"`
	ManifestName  string `json:"manifestName"`
	Mode          string `json:"mode"`
	// +optional
	Options *Options `json:"options,omitempty"`
}

type Options struct {
	// +optional
	Cascading bool `json:"cascading"`
	// +optional
	GracePeriodSeconds int `json:"gracePeriodSeconds,omitempty"`
}

// BakeManifest represents a bake manifest stage in Spinnaker.
// NOTE: I suspect this only supports `helm2` style deployments right now.
// NOTE: notifications currently not supported for this stage.
type BakeManifest struct {
	// +optional
	Comments string `json:"comments,omitempty"`
	// +optional
	FailOnFailedExpressions bool `json:"failOnFailedExpressions,omitempty"`
	// +optional
	FailPipeline *bool `json:"failPipeline,omitempty"`
	// +optional
	ContinuePipeline *bool `json:"continuePipeline,omitempty"`
	// +optional
	CompleteOtherBranchesThenFail *bool `json:"completeOtherBranchesThenFail,omitempty"`
	// +optional
	Namespace                   string              `json:"namespace,omitempty"`
	EvaluateOverrideExpressions bool                `json:"evaluateOverrideExpressions,omitempty"`
	ExpectedArtifacts           []Artifact          `json:"expectedArtifacts,omitempty"`
	InputArtifacts              []ArtifactReference `json:"inputArtifacts,omitempty"`
	// +optional
	OutputName string `json:"outputName,omitempty"`
	// +optional
	Overrides map[string]string `json:"overrides,omitempty"`
	// +optional
	RawOverrides     bool   `json:"rawOverrides,omitempty"`
	TemplateRenderer string `json:"templateRenderer,omitempty"`
	// +optional
	RestrictExecutionDuringTimeWindow bool `json:"restrictExecutionDuringTimeWindow,omitempty"`
	// +optional
	RestrictedExecutionWindow `json:"restrictedExecutionWindow,omitempty"`
	// +optional
	SkipWindowText string `json:"skipWindowText,omitempty"`
	// +optional
	StageEnabled `json:"stageEnabled,omitempty"`
}

// Artifact TODO
type Artifact struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	// +optional
	UsePriorArtifact bool `json:"usePriorArtifact,omitempty"`
	// +optional
	UseDefaultArtifact bool `json:"useDefaultArtifact,omitempty"`
	// +optional
	DefaultArtifact *DefaultArtifact `json:"defaultArtifact,omitempty"`
	// +optional
	MatchArtifact `json:"matchArtifact,omitempty"`
}

type DefaultArtifact struct {
	// +optional
	ID string `json:"id,omitempty"`
	// +optional
	ArtifactAccount string `json:"artifactAccount,omitempty"`
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Reference string `json:"reference,omitempty"`
	// +optional
	Type string `json:"type,omitempty"`
	// +optional
	Version string `json:"version,omitempty"`
}

type MatchArtifact struct {
	// +optional
	ID string `json:"id,omitempty"`
	// +optional
	ArtifactAccount string `json:"artifactAccount,omitempty"`
	// +optional
	Reference string `json:"string,omitempty"`
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Type string `json:"type,omitempty"`
	// +optional
	Version string `json:"version,omitempty"`
}

// TodoArtifact represents an artifact in Spinnaker. TODO also a candidate for union type
type TodoArtifact struct {
	ArtifactAccount string `json:"artifactAccount"` // TODO should be enum eventually
	// +optional
	CustomKind bool   `json:"customKind,omitempty"`
	Location   string `json:"location,omitempty"`
	Name       string `json:"name,omitempty"`
	Reference  string `json:"reference,omitempty"`
	Type       string `json:"type"`
	Version    string `json:"version,omitempty"`
}

// ArtifactReference TODO doesn't seem to be working...?
type ArtifactReference struct {
	Account string `json:"account"`
	ID      string `json:"id"`
}

// FindArtifactsFromResource represents the stage of the same name in Spinnaker.
type FindArtifactsFromResource struct {
	Account       string `json:"account,omitempty"`
	App           string `json:"app,omitempty"`
	CloudProvider string `json:"cloudProvider,omitempty"`
	Location      string `json:"location,omitempty"`
	ManifestName  string `json:"manifestName,omitempty"`
	Mode          string `json:"mode,omitempty"` // FIXME enum static/dynamic
}

// StageEnabled TODO this will need to change
type StageEnabled struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

// ManualJudgment TODO description
type ManualJudgment struct {
	Name                              string           `json:"name,omitempty"`
	Comments                          string           `json:"comments,omitempty"`
	FailPipeline                      bool             `json:"failPipeline,omitempty"`
	Instructions                      string           `json:"instructions,omitempty"`
	JudgmentInputs                    *[]JudgmentInput `json:"judgmentInputs,omitempty"` // No, the json annotation is not spelled incorrectly.
	RestrictExecutionDuringTimeWindow bool             `json:"restrictExecutionDuringTimeWindow,omitempty"`
	RestrictedExecutionWindow         `json:"restrictedExecutionWindow,omitempty"`
	SkipWindowText                    string `json:"skipWindowText,omitempty"`
	StageTimeoutMs                    int    `json:"stageTimeoutMs,omitempty"`
	// +optional
	SendNotifications bool `json:"sendNotifications,omitempty"`
	// +optional
	Notifications []ManualJudgmentNotification `json:"notifications,omitempty"`
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

// ToSpinnakerStage TODO description
func (su StageUnion) ToSpinnakerStage() (map[string]interface{}, error) {
	v := reflect.Indirect(reflect.ValueOf(&su))
	crdType := v.FieldByName("Type").String()
	crdStage := v.FieldByName(crdType).Interface() // TODO this causes a panic, fix it

	var mapified map[string]interface{}

	switch crdType {
	case "BakeManifest":
		crd := crdStage.(BakeManifest)
		// If this value is lowercase the Spinnaker API apparently discards it.
		crd.TemplateRenderer = strings.ToUpper(crd.TemplateRenderer)

		s := structs.New(crd)
		s.TagName = "json"
		mapified = s.Map()
	case "FindArtifactsFromResource":
		s := structs.New(crdStage.(FindArtifactsFromResource))
		s.TagName = "json"
		mapified = s.Map()
	case "ManualJudgment":
		s := structs.New(crdStage.(ManualJudgment))
		s.TagName = "json"
		mapified = s.Map()
	case "DeleteManifest":
		s := structs.New(crdStage.(DeleteManifest))
		s.TagName = "json"
		mapified = s.Map()
	case "CheckPreconditions":
		s := structs.New(crdStage.(CheckPreconditions))
		s.TagName = "json"
		mapified = s.Map()
	case "DeployManifest":
		s := structs.New(crdStage.(DeployManifest))
		s.TagName = "json"
		mapified = s.Map()

		manifests := mapified["manifests"].([]string)
		if len(manifests) > 0 {
			var finalManifests []map[string]interface{}

			for _, stringManifest := range manifests {
				manifest := make(map[string]interface{})
				err := yaml.Unmarshal([]byte(stringManifest), manifest)
				if err != nil {
					return mapified, err
				}
				finalManifests = append(finalManifests, manifest)
			}
			mapified["manifests"] = finalManifests
		}
	}

	if mapified == nil {
		return mapified, fmt.Errorf("could not serialize stage")
	}

	// These fields belong to the top level StageUnion and need to be
	// persisted to the final map[string]interface{} that `plank` uses.
	mapified["type"] = strcase.ToLowerCamel(crdType)
	mapified["name"] = su.Name
	mapified["refId"] = su.RefID
	mapified["requisiteStageRefIds"] = su.RequisiteStageRefIds
	if su.StageEnabled != nil {
		s := structs.New(su.StageEnabled)
		s.TagName = "json"
		mapified["stageEnabled"] = s.Map()
	}

	return mapified, nil
}
