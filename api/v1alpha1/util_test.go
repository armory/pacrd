package v1alpha1

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestGenerateManifestNameHappyPath(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("GenerateManifestName should always create a manifest name if well defined", prop.ForAll(
		func(kind, target string) bool {
			m := make(map[string]interface{})
			m["kind"] = kind
			m["targetName"] = target

			manifestName, err := GenerateManifestName(m)

			return err == nil && manifestName == (kind+" "+target)
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestGenerateManifestNameError(t *testing.T) {
	var m map[string]interface{}
	_, err := GenerateManifestName(m)

	if err == nil {
		t.Fatalf("GenerateManifestName should have errored")
	}
}

func CreateRawExtension(obj interface{}) runtime.RawExtension {
	b, _ := json.Marshal(obj)
	return runtime.RawExtension{Raw: b}
}
