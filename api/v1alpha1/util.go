package v1alpha1

import (
	"encoding/json"
	"fmt"
)

// GenerateManifestName creates a manifest name for use in delete/undo manifest stages.
func GenerateManifestName(m map[string]interface{}) (string, error) {
	if m["kind"] != nil && m["targetName"] != nil {
		return fmt.Sprintf("%v %v", m["kind"], m["targetName"]), nil
	}

	return "", fmt.Errorf("could not construct manifestName, kind or targetName was empty")
}

func stringToMapInterface(stringToConvert string) (map[string]interface{}, error) {
	valuesMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(stringToConvert), &valuesMap)
	return valuesMap, err
}

func rewriteStringValueFromMapToMapInterface(field string, mapified map[string]interface{}) error {
	if fieldString, ok := mapified[field].(string); ok {
		payloadMap, err := stringToMapInterface(fieldString)
		if err != nil {
			return err
		}
		mapified[field] = payloadMap
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}
