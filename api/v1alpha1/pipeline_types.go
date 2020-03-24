/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/armory/plank"
	"github.com/fatih/structs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PipelineSpec defines the desired state of Pipeline
type PipelineSpec struct {
	// Application is a reference to the application that owns this pipeline.
	Application string `json:"application,omitempty"`
	// Description tells the user what this pipeline is for.
	Description string `json:"description,omitempty"`
	// +optional
	ParameterConfig *[]Parameter `json:"parameterConfig,omitempty"`
	// +optional
	ExpectedArtifacts *[]Artifact `json:"expectedArtifacts,omitempty"`
	// ExecutionEngine TODO
	ExecutionEngine string `json:"executionEngine,omitempty"`
	// AllowParallelExecutions TODO
	AllowParallelExecutions bool `json:"allowParallelExecutions,omitempty"`
	// LimitConcurrent TODO
	LimitConcurrent bool `json:"limitConccurent,omitempty"`
	// KeepWaitingPipelines TODO
	KeepWaitingPipelines bool `json:"keepWaitingPipelines,omitempty"`
	// Stages TODO
	Stages []StageUnion `json:"stages"`
}

type Parameter struct {
	// +optional
	Default string `json:"default,omitempty"`
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	HasOptions bool `json:"hasOptions,omitempty"`
	// +optional
	Label string `json:"label,omitempty"`
	Name  string `json:"name"`
	// +optional
	Options *[]OptionValue `json:"options,omitempty"`
	// +optional
	Pinned bool `json:"pinned,omitempty"`
	// +optional
	Required bool `json:"required,omitempty"`
}

type OptionValue struct {
	Value string `json:"value"`
}

// PipelinePhase represents the various stages a pipeline could be in with Spinnaker.
//+kubebuilder:validation:Enum=ErrorNotFound;Creating;ErrorFailedToCreate;Created;Deleting;ErrorDeletingPipeline;Updated;ErrorUpdatingPipeline
type PipelinePhase string

const (
	// PipelineOrApplicationNotFound means a pipeline couldn't be attached to an application.
	PipelineOrApplicationNotFound PipelinePhase = "ErrorNotFound"
	// PipelineCreating means the pipeline is being created in Spinnaker
	PipelineCreating PipelinePhase = "Creating"
	// PipelineCreationFailed indicates that the pipeline could not be saved upstream.
	PipelineCreationFailed PipelinePhase = "ErrorFailedToCreate"
	// PipelineCreated indicates the pipeline was successfully saved upstream.
	PipelineCreated PipelinePhase = "Created"
	// PipelineDeleting indicates the pipeline is in the process of deing deleted upstream.
	PipelineDeleting PipelinePhase = "Deleting"
	// PipelineDeletionFailed indicates there was a problem deleting the pipeline upstream.
	PipelineDeletionFailed PipelinePhase = "ErrorDeletingPipeline"
	// PipelineUpdateFailed indicates that a pipeline failed to update upstream.
	PipelineUpdateFailed PipelinePhase = "ErrorUpdatingPipeline"
	// PipelineUpdated indicates that a pipeline was successfully updated upstream.
	PipelineUpdated PipelinePhase = "Updated"
)

// +kubebuilder:object:root=true

// Pipeline is the Schema for the pipelines API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="status",type="string",JSONPath=".status.phase",description="Status"
// +kubebuilder:printcolumn:name="lastConfigured",type="date",JSONPath=".status.lastConfigured",description="Last Configured"
// +kubebuilder:printcolumn:JSONPath=".status.url",name=URL,type=string
// +kubebuilder:resource:path=pipelines,shortName=pipe
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

// PipelineStatus defines the observed state of Pipeline
type PipelineStatus struct {
	// LastConfigured represents the last time the operator updated this pipeline in Spinnaker.
	LastConfigured metav1.Time `json:"lastConfigured,omitempty"`
	// Phase is the current phase of pipeline reconciliation.
	Phase PipelinePhase `json:"phase,omitempty"`
	URL   string        `json:"url,omitempty"`
	// ID represents the Spinnaker generated id for this pipeline
	ID string `json:"id,omitempty"`
}

// +kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}

// ToSpinnakerPipeline converts a CRD to a value that can be sent to Spinnaker.
func (p Pipeline) ToSpinnakerPipeline() (plank.Pipeline, error) {
	plankPipe := plank.Pipeline{
		ID:                   p.Status.ID,
		Name:                 p.GetObjectMeta().GetName(),
		Application:          p.Spec.Application,
		Description:          p.Spec.Description,
		ExecutionEngine:      p.Spec.ExecutionEngine,
		Parallel:             p.Spec.AllowParallelExecutions,
		LimitConcurrent:      p.Spec.LimitConcurrent,
		KeepWaitingPipelines: p.Spec.KeepWaitingPipelines,
	}

	if p.Spec.ExpectedArtifacts != nil {
		for _, e := range *p.Spec.ExpectedArtifacts {
			s := structs.New(e)
			s.TagName = "json"
			plankPipe.ExpectedArtifacts = append(plankPipe.ExpectedArtifacts, s.Map())
		}
	}

	if p.Spec.ParameterConfig != nil {

		for _, p := range *p.Spec.ParameterConfig {
			s := structs.New(p)
			s.TagName = "json"
			plankPipe.Parameters = append(plankPipe.Parameters, s.Map())
		}
	}

	// TODO this doesn't work yet
	for _, s := range p.Spec.Stages {
		renderedStage, err := s.ToSpinnakerStage()
		if err != nil {
			return plank.Pipeline{}, err
		}
		plankPipe.Stages = append(plankPipe.Stages, renderedStage)
	}

	return plankPipe, nil
}

// ShouldDelete tells you if the current pipeline should be deleted or not.
// NOTE: This assumes finalizers have been set appropriately. If it's empty then this will never be true.
func (p Pipeline) ShouldDelete() bool {
	return !p.GetObjectMeta().GetDeletionTimestamp().IsZero()
}
