package v1alpha1

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime"
)

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

func (su *MatchStage) MarshallToMap() map[string]interface{} {
	s := structs.New(su)
	s.TagName = "json"
	return s.Map()
}

func (su *MatchStage) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, su)

	if err != nil {
		return err
	}

	return nil
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

// BakeManifest represents a bake manifest stage in Spinnaker.
// NOTE: I suspect this only supports `helm2` style deployments right now.
// NOTE: notifications currently not supported for this stage.
type BakeManifest struct {
	Stage `json:",inline"`

	// Name is the name given to this stage.
	Type string `json:"type"`

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

func (bk *BakeManifest) MarshallToMap() map[string]interface{} {
	s := structs.New(bk)
	s.TagName = "json"
	stage := s.Map()
	m := structs.New(bk.Stage)
	m.TagName = "json"
	for key, element := range m.Map() {
		stage[key] = element
	}
	return stage
}

func (bk *BakeManifest) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, bk)

	if len(bk.Stage.RequisiteStageRefIds) == 0 {
		bk.Stage.RequisiteStageRefIds = []string{}
	}

	if len(bk.Stage.RequisiteStageRefIds) == 0 {
		bk.Stage.RequisiteStageRefIds = []string{}
	}

	if len(bk.InputArtifacts) == 0 {
		bk.InputArtifacts = []*ArtifactReference{}
	}

	if len(bk.Overrides) == 0 {
		bk.Overrides = make(map[string]string)
	}

	if err != nil {
		return err
	}

	return nil
}

type CheckPreconditions struct {
	Stage `json:",inline"`
	Type  string `json:"type"`
	// +optional
	Preconditions []*Precondition `json:"preconditions"`
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

func (cp *CheckPreconditions) MarshallToMap() map[string]interface{} {
	s := structs.New(cp)
	s.TagName = "json"
	stage := s.Map()
	m := structs.New(cp.Stage)
	m.TagName = "json"
	for key, element := range m.Map() {
		stage[key] = element
	}
	return stage
}

func (cp *CheckPreconditions) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, cp)

	if err != nil {
		return err
	}

	return nil
}

type DeleteManifest struct {
	Stage         `json:",inline"`
	Type          string `json:"type"`
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

type LabelSelector struct {
	Selector []Selector `json:"selectors,omitempty"`
}
type Options struct {
	// +optional
	Cascading bool `json:"cascading"`
	// +optional
	GracePeriodSeconds *int `json:"gracePeriodSeconds"`
}

//This value comes from: /clouddriver/clouddriver-kubernetes-v2/src/main/java/com/netflix/spinnaker/clouddriver/kubernetes/v2/security/KubernetesSelector.java
// +kubebuilder:validation:Enum=ANY;EQUALS;NOT_EQUALS;CONTAINS;NOT_CONTAINS;EXISTS;NOT_EXISTS
type SelectorsKind string

type Selector struct {
	Key           string `json:"key,omitempty"`
	SelectorsKind `json:"kind,omitempty"`
	Values        []string `json:"values,omitempty"`
}

func (dm *DeleteManifest) MarshallToMap() map[string]interface{} {
	s := structs.New(dm)
	s.TagName = "json"
	stage := s.Map()
	m := structs.New(dm.Stage)
	m.TagName = "json"
	for key, element := range m.Map() {
		stage[key] = element
	}

	//When we have static target the manifestname is the union of kind and targetName
	if modevalue, ok := stage["mode"]; ok && modevalue == ChooseStaticTarget {
		manifestName, err := GenerateManifestName(stage)

		if err != nil {
			return stage
		}

		stage["manifestName"] = manifestName
	}

	return stage
}

func (dm *DeleteManifest) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, dm)

	if err != nil {
		return err
	}

	return nil
}

// DeployManifest deploys a Kubernetes manifest to a target Kubernetes cluster. Spinnaker will periodically check the status of the manifest to make sure the manifest converges on the target cluster until it reaches a timeout
type DeployManifest struct {
	Stage `json:",inline"`
	Type  string `json:"type"`
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

func (dm *DeployManifest) MarshallToMap() map[string]interface{} {
	s := structs.New(dm)
	s.TagName = "json"
	stage := s.Map()
	m := structs.New(dm.Stage)
	m.TagName = "json"
	for key, element := range m.Map() {
		stage[key] = element
	}

	if _, ok := stage["manifests"]; ok {
		manifests := stage["manifests"].([]string)
		if len(manifests) > 0 {
			var finalManifests []map[string]interface{}

			for _, stringManifest := range manifests {
				manifest := make(map[string]interface{})
				err := yaml.Unmarshal([]byte(stringManifest), &manifest)
				if err != nil {
					return stage
				}
				finalManifests = append(finalManifests, manifest)
			}
			stage["manifests"] = finalManifests
		}
	}
	return stage
}

func (dm *DeployManifest) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, dm)

	if err != nil {
		return err
	}

	return nil
}

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

func (fafr *FindArtifactsFromResource) MarshallToMap() map[string]interface{} {
	s := structs.New(fafr)
	s.TagName = "json"
	stage := s.Map()
	m := structs.New(fafr.Stage)
	m.TagName = "json"
	for key, element := range m.Map() {
		stage[key] = element
	}
	return stage
}

func (fafr *FindArtifactsFromResource) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, fafr)

	if err != nil {
		return err
	}

	if fafr.Account == "" {
		return fmt.Errorf("account must be defined for this stage")
	}

	return nil
}

// ManualJudgment TODO description
type ManualJudgment struct {
	Stage          `json:",inline"`
	Type           string `json:"type"`
	FailPipeline   bool   `json:"failPipeline,omitempty"`
	Instructions   string `json:"instructions,omitempty"`
	StageTimeoutMs int    `json:"stageTimeoutMs,omitempty"`
	//+optional
	SendNotifications bool `json:"sendNotifications,omitempty"`
	//+optional
	Notifications []*ManualJudgmentNotification `json:"notifications"`
	//+optional
	JudgmentInputs []JudgmentInput `json:"judgmentInputs"`
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

func (mj *ManualJudgment) MarshallToMap() map[string]interface{} {
	s := structs.New(mj)
	s.TagName = "json"
	stage := s.Map()
	m := structs.New(mj.Stage)
	m.TagName = "json"
	for key, element := range m.Map() {
		stage[key] = element
	}
	return stage
}

func (fafr *ManualJudgment) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, fafr)

	if err != nil {
		return err
	}

	return nil
}

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

func (urm *UndoRolloutManifest) MarshallToMap() map[string]interface{} {
	s := structs.New(urm)
	s.TagName = "json"
	stage := s.Map()
	m := structs.New(urm.Stage)
	m.TagName = "json"
	for key, element := range m.Map() {
		stage[key] = element
	}

	//When we have static target the manifestname is the union of kind and targetName
	if modevalue, ok := stage["mode"]; ok && modevalue == UndoRolloutManifestStaticMode {
		manifestName, err := GenerateManifestName(stage)

		if err != nil {
			return stage
		}

		stage["manifestName"] = manifestName
	}

	return stage
}

func (urm *UndoRolloutManifest) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, urm)

	if err != nil {
		return err
	}

	return nil
}

// +kubebuilder:object:generate=false
type UnknownStage struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

func (us *UnknownStage) MarshallToMap() map[string]interface{} {
	us.Properties["type"] = us.Type // Persist the type we got at ingest time.

	return us.Properties
}

func (us *UnknownStage) NewStageFromBytes(data []byte) error {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	us.Properties = m

	return err
}

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

func (w *Webhook) MarshallToMap() map[string]interface{} {
	s := structs.New(w)
	s.TagName = "json"

	stage := s.Map()
	m := structs.New(w.Stage)
	m.TagName = "json"
	for key, element := range m.Map() {
		stage[key] = element
	}

	err := rewriteStringValueFromMapToMapInterface("payload", stage)
	if err != nil {
		return stage
	}
	err = rewriteStringValueFromMapToMapInterface("cancelPayload", stage)
	if err != nil {
		return stage
	}
	err = rewriteStringValueFromMapToMapInterface("customHeaders", stage)
	if err != nil {
		return stage
	}
	return stage

}

func (w *Webhook) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, w)

	if err != nil {
		return err
	}

	return nil
}

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

// ToSpinnakerStage TODO description
func (su MatchStage) ToSpinnakerStage() (map[string]interface{}, error) {

	s := su.GetStage().(SpinnakerMatchStage)
	err := s.NewStageFromBytes(su.Properties.Raw)

	stage := s.MarshallToMap()
	stage["type"] = su.Type
	delete(stage, "Stage")
	return stage, err
}
