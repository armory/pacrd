package v1alpha1

import (
	"encoding/json"
	"github.com/fatih/structs"
	"github.com/ghodss/yaml"
)

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

func (dm *DeployManifest) NewStageFromBytes(data []byte) error {
	err := json.Unmarshal(data, dm)

	if err != nil {
		log.WithName("DeployManifest").Error(err, "error while reading DeployManifest")
		return err
	}

	return nil
}

func (dm *DeployManifest) MarshallToMap() map[string]interface{} {
	s := structs.New(dm)
	s.TagName = "json"
	stage := s.Map()
	for key, element := range dm.Stage.MarshallToMap() {
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
					log.WithName("DeployManifest").Error(err, "error while trying to read manifest content")
					return stage
				}
				finalManifests = append(finalManifests, manifest)
			}
			stage["manifests"] = finalManifests
		}
	}
	return stage
}
