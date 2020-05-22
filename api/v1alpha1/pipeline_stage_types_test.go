package v1alpha1

import (
	"encoding/json"
	"fmt"
	"gotest.tools/assert"
	"testing"
)

func TestStageUnion_ToSpinnakerStage_vs_Deck(t *testing.T) {
	tests := map[string]struct {
		stage    StageUnion
		deckJson string
	}{
		"bakeManifest": {
			stage: StageUnion{
				Name:         "Bake (Manifest)",
				RefID:        "1",
				BakeManifest: BakeManifest{},
				Type:         "BakeManifest",
			},
			deckJson: `{
     "inputArtifacts": [],
     "name": "Bake (Manifest)",
     "overrides": {},
     "refId": "1",
     "requisiteStageRefIds": [],
     "type": "bakeManifest"
    }`,
		},
		"findArtifactsFromResource": {
			stage: StageUnion{
				Name:  "Find Artifacts From Resource (Manifest)",
				RefID: "1",
				FindArtifactsFromResource: FindArtifactsFromResource{
					Account:       "kubernetes",
					App:           "decktesting",
					CloudProvider: "kubernetes",
					Location:      "test",
					Mode:          "static",
				},
				Type: "FindArtifactsFromResource",
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
			stage: StageUnion{
				Name:           "Manual Judgment",
				RefID:          "1",
				ManualJudgment: ManualJudgment{},
				Type:           "ManualJudgment",
			},
			deckJson: `{
      "failPipeline": true,
      "judgmentInputs": [],
      "name": "Manual Judgment",
      "notifications": [],
      "refId": "1",
      "requisiteStageRefIds": [],
      "type": "manualJudgment"
    }`,
		},
		"deleteManifest": {
			stage: StageUnion{
				Name:  "Delete (Manifest)",
				RefID: "1",
				DeleteManifest: DeleteManifest{
					Account:            "kubernetes",
					App:                "decktesting",
					CloudProvider:      "kubernetes",
					Location:           "test",
					DeleteManifestMode: "static",
					KubernetesKind:     ApiService,
					TargetName:         "testing",
				},
				Type: "DeleteManifest",
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
        "cascading": false,
        "gracePeriodSeconds": null
      },
      "refId": "1",
      "requisiteStageRefIds": [],
      "type": "deleteManifest"
    }`,
		},
		"undoRolloutManifest": {
			stage: StageUnion{
				Name:  "Undo Rollout (Manifest)",
				RefID: "1",
				UndoRolloutManifest: UndoRolloutManifest{
					Account:          "kubernetes",
					CloudProvider:    "kubernetes",
					Location:         "test",
					Mode:             "static",
					TargetName:       "testing",
					Kind:             ApiService,
					NumRevisionsBack: 1,
				},
				Type: "UndoRolloutManifest",
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
			stage: StageUnion{
				Name:               "Check Preconditions",
				RefID:              "1",
				CheckPreconditions: CheckPreconditions{},
				Type:               "CheckPreconditions",
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
			stage: StageUnion{
				Name:  "Webhook",
				RefID: "1",
				Webhook: Webhook{
					StatusUrlResolution: "getMethod",
				},
				Type: "Webhook",
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
