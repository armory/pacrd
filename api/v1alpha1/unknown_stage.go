package v1alpha1

import "encoding/json"

// +kubebuilder:object:generate=false
type UnknownStage struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

func (us *UnknownStage) NewStageFromBytes(data []byte) error {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	if err != nil {
		log.WithName("UnknownStage").Error(err, "error while reading Stage")
		return err
	}

	us.Properties = m

	return err
}

func (us *UnknownStage) MarshallToMap() map[string]interface{} {
	us.Properties["type"] = us.Type // Persist the type we got at ingest time.

	return us.Properties
}
