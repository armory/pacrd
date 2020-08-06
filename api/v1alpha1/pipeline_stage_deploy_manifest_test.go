package v1alpha1

import (
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"testing"
)

func TestToSpinnakerStage_DeployManifest_HappyPath(t *testing.T) {
	properties := gopter.NewProperties(nil)

	sampleYaml := `
apiVersion: v1
kind: Deployment
metadata:
  name: foo
spec:
  containers:
    - name: bar
      image: nginx:1.17
`

	properties.Property("ToSpinnakerStage should serialize a deploy manifest stage given non-empty inputs", prop.ForAll(
		func(name, account, cloudProvider string, refId int, completeThenFail, failPipeline bool) bool {
			// A PaCRD stage minimizing the # of variables under test
			deployManifest := DeployManifest{
				Account:       account,
				CloudProvider: cloudProvider,
				Manifests:     []string{sampleYaml},
				Stage: Stage{
					Name:  name,
					RefID: string(refId),
				},
			}

			stage := MatchStage{
				Type:       "deployManifest",
				Properties: CreateRawExtension(deployManifest),
			}

			// The method under test
			mapStage, err := stage.ToSpinnakerStage()

			// Validate that no error is thrown in serialization
			if err != nil {
				t.Logf("%+v", err)
				return false
			}

			// Ensure that the manifest has been serialized correctly
			if mapStage["manifests"] != nil {
				s := mapStage["manifests"].([]map[string]interface{})
				kind := s[0]["kind"].(string)
				metadata := s[0]["metadata"].(map[string]interface{})
				return kind == "Deployment" && metadata["name"].(string) == "foo"
			}

			return false
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int(),
		gen.Bool(),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

func TestToSpinnakerStage_DeployManifest_ArtifactHappyPath(t *testing.T) {

	properties := gopter.NewProperties(nil)

	properties.Property("ToSpinnakerStage should serialize a deploy manifest stage without inline manifests", prop.ForAll(
		func(name, account, cloudProvider, artifactId string, refId int, completeThenFail, failPipeline bool) bool {
			// A PaCRD stage minimizing the # of variables under test
			deployManifest := DeployManifest{
				Account:                 account,
				CloudProvider:           cloudProvider,
				ManifestArtifactAccount: account,
				ManifestArtifactID:      artifactId,
				Stage: Stage{
					Name:  name,
					RefID: string(refId),
				},
			}

			stage := MatchStage{
				Type:       "deployManifest",
				Properties: CreateRawExtension(deployManifest),
			}

			// The method under test
			mapStage, err := stage.ToSpinnakerStage()

			// Validate that no error is thrown in serialization
			if err != nil {
				t.Logf("%+v", err)
				return false
			}

			// Ensure that the manifest has been serialized correctly
			return mapStage["manifestArtifactAccount"].(string) == account && mapStage["manifestArtifactId"] == artifactId && mapStage["manifests"] == nil
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int(),
		gen.Bool(),
		gen.Bool(),
	))
}
