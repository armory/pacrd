package v1alpha1

// UndoRolloutManifestMode is the means for undoing a manifest rollout.
// +kubebuilder:validation:Enum=static
type UndoRolloutManifestMode string

const (
	// UndoRolloutManifestStaticMode .
	UndoRolloutManifestStaticMode UndoRolloutManifestMode = "static"
)

// UndoRolloutManifest is a stage that rolls back a manifest.
type UndoRolloutManifest struct {
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
