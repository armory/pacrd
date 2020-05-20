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
// +kubebuilder:validation:Enum=BakeManifest;FindArtifactsFromResource;ManualJudgment;DeleteManifest;CheckPreconditions;DeployManifest;Webhook;UndoRolloutManifest
type StageUnionType string

// Stage is a union type that encompasses strongly typed stage defnitions.
type StageUnion struct {
	// Type represents the type of stage that is described.
	Type StageUnionType `json:"type"`
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
	RestrictedExecutionWindow `json:"restrictedExecutionWindow"`
	// SkipWindowText is the text to display when this stage is skipped.
	// +optional
	SkipWindowText string `json:"skipWindowText"`
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
	//Webhook allows you to make quick API calls to an external system as part of a pipeline
	// +optional
	Webhook `json:"webhook,omitempty"`
	// UndoRolloutManifest rolls back a Kubernetes manifest to a previous version.
	//+optional
	UndoRolloutManifest `json:"undoRolloutManifest,omitempty"`
}

// DeployManifest deploys a Kubernetes manifest to a target Kubernetes cluster. Spinnaker will periodically check the status of the manifest to make sure the manifest converges on the target cluster until it reaches a timeout
type DeployManifest struct {
	// Account is the configured account to deploy to.
	Account string `json:"account"`
	// CloudProvider is the type of cloud provider used by the selected account.
	CloudProvider string `json:"cloudProvider"`
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
	// +optional
	ManifestArtifact *MatchArtifact `json:"manifestArtifact,omitempty"`
	// +optional
	NamespaceOverride string `json:"namespaceOverride,omitempty"`
	// +optional
	RequiredArtifacts []Artifact `json:"requiredArtifacts,omitempty"`
	// +optional
	RequiredArtifactIds []string `json:"requiredArtifactIds,omitempty"`
	// Spinnaker manages traffic based on your selected strategy
	// +optional
	TrafficManagement `json:"trafficManagement,omitempty"`
	// +optional
	StageTimeoutMs int `json:"stageTimeoutMs,omitempty"`
}

// Spinnaker manages traffic based on your selected strategy
type TrafficManagement struct {
	// Allow Spinnaker to associate each ReplicaSet deployed in this stage with one or more Services
	// and manage traffic based on your selected rollout strategy options.
	Enabled bool `json:"enabled,omitempty"`
	// +optional
	TrafficManagementOptions `json:"options,omitempty"`
}

// TrafficManagementOptions
type TrafficManagementOptions struct {
	// Send client requests to new pods
	EnableTraffic             bool     `json:"enableTraffic,omitempty"`
	Namespace                 string   `json:"namespace,omitempty"`
	Services                  []string `json:"services,omitempty"`
	TrafficManagementStrategy `json:"strategy,omitempty"`
}

// Tells Spinnaker what to do with the previous version(s) of the ReplicaSet in the cluster.
// Redblack: Disables all previous ReplicaSet as soon as the new ReplicaSet is ready.
// Highlander: Destroys all previous ReplicaSet as soon as the new ReplicaSet is ready.
// +kubebuilder:validation:Enum=redblack;highlander
type TrafficManagementStrategy string

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

//This value comes from: /clouddriver/clouddriver-kubernetes-v2/src/main/java/com/netflix/spinnaker/clouddriver/kubernetes/v2/security/KubernetesSelector.java
// +kubebuilder:validation:Enum=ANY;EQUALS;NOT_EQUALS;CONTAINS;NOT_CONTAINS;EXISTS;NOT_EXISTS
type SelectorsKind string

type Selector struct {
	Key           string `json:"key,omitempty"`
	SelectorsKind `json:"kind,omitempty"`
	Values        []string `json:"values,omitempty"`
}

type LabelSelector struct {
	Selector []Selector `json:"selectors,omitempty"`
}

type DeleteManifest struct {
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
	FailOnFailedExpressions bool `json:"failOnFailedExpressions,omitempty"`
	// +optional
	FailPipeline *bool `json:"failPipeline,omitempty"`
	// +optional
	ContinuePipeline *bool `json:"continuePipeline,omitempty"`
	// +optional
	CompleteOtherBranchesThenFail *bool `json:"completeOtherBranchesThenFail,omitempty"`
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

// FindArtifactsFromResource represents the stage of the same name in Spinnaker.
type FindArtifactsFromResource struct {
	Account       string `json:"account,omitempty"`
	App           string `json:"app,omitempty"`
	CloudProvider string `json:"cloudProvider,omitempty"`
	Location      string `json:"location,omitempty"`
	ManifestName  string `json:"manifestName,omitempty"`
	Mode          string `json:"mode,omitempty"` // FIXME enum static/dynamic
}

// StageEnabled represents whether this stage is active in a pipeline graph.
type StageEnabled struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

// ManualJudgment TODO description
type ManualJudgment struct {
	Name           string `json:"name,omitempty"`
	FailPipeline   bool   `json:"failPipeline,omitempty"`
	Instructions   string `json:"instructions,omitempty"`
	StageTimeoutMs int    `json:"stageTimeoutMs,omitempty"`
	// +optional
	SendNotifications bool `json:"sendNotifications,omitempty"`
	// +optional
	Notifications []ManualJudgmentNotification `json:"notifications"`
	// +optional
	JudgmentInputs []JudgmentInput `json:"judgmentInputs"` // No, the json annotation is not spelled incorrectly.
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

// StatusUrlResolution will poll a status url to determine the progress of the stage.
// +kubebuilder:validation:Enum=locationHeader;getMethod;webhookResponse
type StatusUrlResolution string

// Webhook represents a webhook stage in Spinnaker.
// NOTE: notifications currently not supported for this stage.
type Webhook struct {
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

		//When overrides is not present we need to sent it anyways, otherwise rosco fails
		if overrideval, ok := mapified["overrides"]; !ok || overrideval == nil {
			mapified["overrides"] = map[string]string{}
		}

		var mapifiedArtifacts []map[string]interface{}
		for _, a := range crd.ExpectedArtifacts {
			artifact, err := a.MarshallToMap()

			if err != nil {
				return map[string]interface{}{}, err // TODO wrap error
			}

			mapifiedArtifacts = append(mapifiedArtifacts, artifact)
		}
		mapified["expectedArtifacts"] = mapifiedArtifacts
	case "FindArtifactsFromResource":
		s := structs.New(crdStage.(FindArtifactsFromResource))
		s.TagName = "json"
		mapified = s.Map()
	case "ManualJudgment":
		s := structs.New(crdStage.(ManualJudgment))
		s.TagName = "json"
		mapified = s.Map()

		if _, ok := mapified["failPipeline"]; !ok {
			mapified["failPipeline"] = true
		}
	case "DeleteManifest":
		s := structs.New(crdStage.(DeleteManifest))
		s.TagName = "json"
		mapified = s.Map()

		//When we have static target the manifestname is the union of kind and targetName
		if modevalue, ok := mapified["mode"]; ok && modevalue == ChooseStaticTarget {
			manifestName, err := GenerateManifestName(mapified)

			if err != nil {
				return mapified, err
			}

			mapified["manifestName"] = manifestName
		}
	case "UndoRolloutManifest":
		s := structs.New(crdStage.(UndoRolloutManifest))
		s.TagName = "json"
		mapified = s.Map()

		//When we have static target the manifestname is the union of kind and targetName
		if modevalue, ok := mapified["mode"]; ok && modevalue == UndoRolloutManifestStaticMode {
			manifestName, err := GenerateManifestName(mapified)

			if err != nil {
				return mapified, err
			}

			mapified["manifestName"] = manifestName
		}
	case "CheckPreconditions":
		s := structs.New(crdStage.(CheckPreconditions))
		s.TagName = "json"
		mapified = s.Map()
	case "DeployManifest":
		s := structs.New(crdStage.(DeployManifest))
		s.TagName = "json"
		mapified = s.Map()

		if _, ok := mapified["manifests"]; ok {
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
	case "Webhook":
		s := structs.New(crdStage.(Webhook))
		s.TagName = "json"
		mapified = s.Map()

		err := rewriteStringValueFromMapToMapInterface("payload", mapified)
		if err != nil {
			return mapified, err
		}
		err = rewriteStringValueFromMapToMapInterface("cancelPayload", mapified)
		if err != nil {
			return mapified, err
		}
		err = rewriteStringValueFromMapToMapInterface("customHeaders", mapified)
		if err != nil {
			return mapified, err
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
	mapified["comments"] = su.Comments
	mapified["restrictExecutionDuringTimeWindow"] = su.RestrictExecutionDuringTimeWindow
	mapified["restrictedExecutionWindow"] = su.RestrictedExecutionWindow
	mapified["skipWindowText"] = su.SkipWindowText
	if su.StageEnabled != nil {
		s := structs.New(su.StageEnabled)
		s.TagName = "json"
		mapified["stageEnabled"] = s.Map()
	}
	// Always should be empty, not nil
	if su.RequisiteStageRefIds == nil {
		mapified["requisiteStageRefIds"] = []string{}
	} else {
		mapified["requisiteStageRefIds"] = su.RequisiteStageRefIds
	}

	return mapified, nil
}

func rewriteStringValueFromMapToMapInterface(field string, mapified map[string]interface{}) error {
	if fieldString, ok := mapified[field].(string); ok {
		payloadMap, err := stringToMapInterface(fieldString)
		if err != nil {
			return err
		}
		mapified[field] = payloadMap
	}
	return nil
}

func stringToMapInterface(stringToConvert string) (map[string]interface{}, error) {
	valuesMap := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(stringToConvert), valuesMap)
	return valuesMap, err
}
