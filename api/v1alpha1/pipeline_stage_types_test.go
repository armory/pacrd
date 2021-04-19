package v1alpha1

import (
	"encoding/json"
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestMatchStage_ToSpinnakerStage_vs_Deck(t *testing.T) {
	bakeManifest := BakeManifest{
		Stage: Stage{
			Name:  "Bake (Manifest)",
			RefID: "1",
		},
		Namespace: "default",
	}

	findArtifactsFromResource := FindArtifactsFromResource{
		Stage: Stage{
			Name:  "Find Artifacts From Resource (Manifest)",
			RefID: "1",
		},
		Account:       "kubernetes",
		App:           "decktesting",
		CloudProvider: "kubernetes",
		Location:      "test",
		Mode:          "static",
	}

	manualJudgment := ManualJudgment{
		Stage: Stage{
			Name:  "Manual Judgment",
			RefID: "1",
		},
	}

	deleteManifest := DeleteManifest{
		Stage: Stage{
			Name:  "Delete (Manifest)",
			RefID: "1",
		},
		Account:            "kubernetes",
		App:                "decktesting",
		CloudProvider:      "kubernetes",
		Location:           "test",
		DeleteManifestMode: "static",
		KubernetesKind:     ApiService,
		TargetName:         "testing",
		Options: &Options{
			Cascading:          true,
			GracePeriodSeconds: nil,
		},
	}

	undoRolloutManifest := UndoRolloutManifest{
		Stage: Stage{
			Name:  "Undo Rollout (Manifest)",
			RefID: "1",
		},
		Account:          "kubernetes",
		CloudProvider:    "kubernetes",
		Location:         "test",
		Mode:             "static",
		TargetName:       "testing",
		Kind:             ApiService,
		NumRevisionsBack: 1,
	}

	checkPreconditions := CheckPreconditions{
		Stage: Stage{
			Name:  "Check Preconditions",
			RefID: "1",
		},
	}

	webhook := Webhook{
		Stage: Stage{
			Name:  "Webhook",
			RefID: "1",
		},
		StatusUrlResolution: "getMethod",
	}

	tests := map[string]struct {
		stage    MatchStage
		deckJson string
	}{
		"bakeManifest": {
			stage: MatchStage{
				Properties: CreateRawExtension(bakeManifest),
				Type:       "bakeManifest",
			},
			deckJson: `{
    "inputArtifacts": [],
	"namespace": "default",
    "name": "Bake (Manifest)",
    "overrides": {},
    "refId": "1",
    "requisiteStageRefIds": [],
    "type": "bakeManifest"
   }`,
		},
		"findArtifactsFromResource": {
			stage: MatchStage{
				Properties: CreateRawExtension(findArtifactsFromResource),
				Type:       "findArtifactsFromResource",
			},
			deckJson: `{
 "account": "kubernetes",
 "app": "decktesting",
 "cloudProvider": "kubernetes",
 "location": "test",
 "manifestName": "",
 "mode": "static",
 "name": "Find Artifacts From Resource (Manifest)",
 "type": "findArtifactsFromResource"
}`,
		},
		"manualJudgment": {
			stage: MatchStage{
				Properties: CreateRawExtension(manualJudgment),
				Type:       "manualJudgment",
			},
			deckJson: `{
     "judgmentInputs": [],
     "name": "Manual Judgment",
     "notifications": [],
     "refId": "1",
     "requisiteStageRefIds": [],
     "type": "manualJudgment"
   }`,
		},
		"deleteManifest": {
			stage: MatchStage{
				Properties: CreateRawExtension(deleteManifest),
				Type:       "deleteManifest",
			},
			deckJson: `{
     "account": "kubernetes",
     "app": "decktesting",
     "cloudProvider": "kubernetes",
     "location": "test",
     "manifestName": "apiService testing",
     "mode": "static",
     "name": "Delete (Manifest)",
     "options": {
       "cascading": true,
       "gracePeriodSeconds": null
     },
     "refId": "1",
     "requisiteStageRefIds": [],
     "type": "deleteManifest"
   }`,
		},
		"undoRolloutManifest": {
			stage: MatchStage{
				Properties: CreateRawExtension(undoRolloutManifest),
				Type:       "undoRolloutManifest",
			},
			deckJson: `{
     "account": "kubernetes",
     "cloudProvider": "kubernetes",
     "location": "test",
     "manifestName": "apiService testing",
     "mode": "static",
     "name": "Undo Rollout (Manifest)",
     "numRevisionsBack": 1,
     "refId": "1",
     "requisiteStageRefIds": [],
     "type": "undoRolloutManifest"
   }`,
		},
		"checkPreconditions": {
			stage: MatchStage{
				Properties: CreateRawExtension(checkPreconditions),
				Type:       "checkPreconditions",
			},
			deckJson: `{
     "name": "Check Preconditions",
     "preconditions": [],
     "refId": "1",
     "requisiteStageRefIds": [],
     "type": "checkPreconditions"
   }`,
		},
		"webhook": {
			stage: MatchStage{
				Properties: CreateRawExtension(webhook),
				Type:       "webhook",
			},
			deckJson: `{
     "name": "Webhook",
     "refId": "1",
     "requisiteStageRefIds": [],
     "statusUrlResolution": "getMethod",
     "type": "webhook"
   }`,
		},
	}
	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			got, errToStage := tt.stage.ToSpinnakerStage()
			if errToStage != nil {
				t.Errorf("Failed to Convert stage to spinnaker stage: %v", errToStage)
				assert.Equal(t, true, false)
			}
			jsonBody, err := json.Marshal(got)
			if err != nil {
				t.Errorf("Failed to marshal stage: %v", err)
				assert.Equal(t, true, false)
			}
			pacMap := make(map[string]interface{})
			_ = json.Unmarshal([]byte(string(jsonBody)), &pacMap)

			deckMap := make(map[string]interface{})
			errMarshal := json.Unmarshal([]byte(tt.deckJson), &deckMap)
			if err != nil {
				t.Errorf("Failed to unmarshal deckJson: %v", errMarshal)
				assert.Equal(t, true, false)
			}

			passed := true
			for k, v := range deckMap {
				gotVal := fmt.Sprintf("%v", got[k])
				vVal := fmt.Sprintf("%v", v)

				if vVal != gotVal {
					t.Errorf("ToSpinnakerStage() got = %v, want %v for %v", gotVal, vVal, k)
					passed = false
				}
			}
			assert.Equal(t, passed, true)
		})
	}
}
