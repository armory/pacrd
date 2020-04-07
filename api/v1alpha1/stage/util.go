package stage

import "fmt"

// GenerateManifestName creates a manifest name for use in delete/undo manifest stages.
func GenerateManifestName(m map[string]interface{}) (string, error) {
	if m["kind"] != nil && m["targetName"] != nil {
		return fmt.Sprintf("%v %v", m["kind"], m["targetName"]), nil
	}

	return "", fmt.Errorf("could not construct manifestName, kind or targetName was empty")
}
