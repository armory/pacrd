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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Permissions maps actions inside Spinnaker to authenticated roles that can take them.
type Permissions struct {
	// +optional
	// Read grants the defined roles the ability to read an application and its pipelines.
	Read *[]string `json:"READ,omitempty"`
	// +optional
	// Write grants the defined roles the ability to modify an application and its pipelines.
	Write *[]string `json:"WRITE,omitempty"`
	// +optional
	// Execute grants the defined roles the ability to execute an application's pipelines.
	Execute *[]string `json:"EXECUTE,omitempty"`
}

// DataSource is a tab in the Spinnaker UI representing a kind of managed resource.
// Allowed values include: serverGroups,executions,loadBalancers,securityGroups.
// +kubebuilder:validation:Enum=serverGroups;executions;loadBalancers;securityGroups
type DataSource string

const (
	// ServerGroup shows instances of an Application.
	ServerGroup DataSource = "serverGroups"
	// Execution shows pipeline status for an Application.
	Execution DataSource = "executions"
	// LoadBalancer shows configured load balancers for an Application.
	LoadBalancer DataSource = "loadBalancers"
	// SecurityGroup shows configured network policies for an Application.
	SecurityGroup DataSource = "securityGroups"
)

// DataSources optionally enable and disable elements of the Spinnaker Application UI.
type DataSources struct {
	// Enabled is the list of explicitly enabled UI elements.
	// +optional
	Enabled *[]DataSource `json:"enabled,omitempty"`
	// Disabled is the list of explicitly disabled UI elements.
	// +optional
	Disabled *[]DataSource `json:"disabled,omitempty"`
}

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// Email points to the e-mail user or list that owns this application.
	Email string `json:"email,omitempty"`
	// Description explains the purpose of this application.
	Description string `json:"description,omitempty"`
	// DataSources optionally enable and disable elements of the Spinnaker Application UI.
	// +optional
	*DataSources `json:"dataSources,omitempty"`
	// Permissions maps actions inside Spinnaker to authenticated roles that can take them.
	// +optional
	*Permissions `json:"permissions,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// LastConfigured represents the last time the operator updated this application in Spinnaker.
	LastConfigured metav1.Time `json:"lastConfigured,omitempty"`
	// Phase represents the current status of this application.
	Phase ApplicationPhase `json:"phase,omitempty"`
	// Url represents the URL of the configured Spinnaker cluster.
	URL string `json:"url,omitempty"`
}

// ApplicationPhase represents the various stages an application could be in with Spinnaker.
//+kubebuilder:validation:Enum=ErrorNotFound;Creating;ErrorFailedToCreate;Created;Deleting;ErrorDeletingApplication;Updated;ErrorUpdatingApplication
type ApplicationPhase string

const (
	// ApplicationOrPipelineNotFound means an application couldn't be attached to an application.
	ApplicationOrPipelineNotFound ApplicationPhase = "ErrorNotFound"
	// ApplicationCreating means the application is being created in Spinnaker
	ApplicationCreating ApplicationPhase = "Creating"
	// ApplicationCreationFailed indicates that the application could not be saved upstream.
	ApplicationCreationFailed ApplicationPhase = "ErrorFailedToCreate"
	// ApplicationCreated indicates the application was successfully saved upstream.
	ApplicationCreated ApplicationPhase = "Created"
	// ApplicationDeleting indicates the application is in the process of deing deleted upstream.
	ApplicationDeleting ApplicationPhase = "Deleting"
	// ApplicationDeletionFailed indicates there was a problem deleting the application upstream.
	ApplicationDeletionFailed ApplicationPhase = "ErrorDeletingApplication"
	// ApplicationUpdateFailed indicates that a application failed to update upstream.
	ApplicationUpdateFailed ApplicationPhase = "ErrorUpdatingApplication"
	// ApplicationUpdated indicates that a application was successfully updated upstream.
	ApplicationUpdated ApplicationPhase = "Updated"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="status",type="string",JSONPath=".status.phase",description="Status"
// +kubebuilder:printcolumn:name="lastConfigured",type="date",JSONPath=".status.lastConfigured",description="Last Configured"
// +kubebuilder:printcolumn:JSONPath=".status.url",name=URL,type=string
// +kubebuilder:resource:path=applications,shortName=app

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}

// ToSpinApplication translates a CRD Application to a Spinnaker Application.
func (a Application) ToSpinApplication() plank.Application {
	spinApp := plank.Application{
		Name:        a.Name,
		Email:       a.Spec.Email,
		Description: a.Spec.Description,
	}

	if a.Spec.DataSources != nil {

		spinApp.DataSources = &plank.DataSourcesType{}

		if a.Spec.DataSources.Disabled != nil {
			var disabledSources []string
			for _, s := range *a.Spec.DataSources.Disabled {
				disabledSources = append(disabledSources, string(s))
			}
			spinApp.DataSources.Disabled = disabledSources
		}
		if a.Spec.DataSources.Enabled != nil {
			var enabledSources []string
			for _, s := range *a.Spec.DataSources.Enabled {
				enabledSources = append(enabledSources, string(s))
			}
			spinApp.DataSources.Enabled = enabledSources
		}
	}

	if a.Spec.Permissions != nil {
		spinApp.Permissions = &plank.PermissionsType{}

		if a.Spec.Permissions.Read != nil {
			spinApp.Permissions.Read = *a.Spec.Permissions.Read
		}

		if a.Spec.Permissions.Write != nil {
			spinApp.Permissions.Write = *a.Spec.Permissions.Write
		}

		if a.Spec.Permissions.Execute != nil {
			spinApp.Permissions.Execute = *a.Spec.Permissions.Execute
		}
	}

	return spinApp
}

// ShouldDelete determines if the current Application should be deleted from Spinnaker.
// NOTE: This assumes finalizers have been set appropriately. If it's empty then this will never be true.
func (a Application) ShouldDelete() bool {
	return !a.GetObjectMeta().GetDeletionTimestamp().IsZero()
}
