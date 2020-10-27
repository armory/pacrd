package spinnaker

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/armory/plank/v3"
)

// Client negotiates interactions with a Spinnaker cluster.
type Client interface {
	GetApplication(string) (*plank.Application, error)
	CreateApplication(*plank.Application) error
	UpdateApplication(plank.Application) error
	DeleteApplication(string) error
	GetPipelines(string) ([]plank.Pipeline, error)
	DeletePipeline(plank.Pipeline) error
	UpsertPipeline(plank.Pipeline, string) error
}

// FrontFiftyBadResponse captures a 4xx error from Front50
// FIXME the more I think about this the more I think this struct and methods
// FIXME should live in `plank`.
type FrontFiftyBadResponse struct {
	Type   string `json:"error"`
	Reason string `json:"message"`
}

func (e *FrontFiftyBadResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Reason)
}

// UnwrapFrontFiftyBadResponse will unwrap a bad response from front50 if it's appropriate, otherwise pass through the existing error.
func UnwrapFrontFiftyBadResponse(e error) error {
	// Front50/OPA errors are in the form {"error": "BadRequest", "message": "foo"}
	// So let's attempt to destructure that value and report it downstream so
	// users have a nicer error message. If we can't parse the message for
	// whatever reason, then we'll just end up returning the previous error instead.
	var fr *plank.FailedResponse
	if errors.As(e, &fr) {
		var ffbr FrontFiftyBadResponse
		if jsonErr := json.Unmarshal(fr.Response, &ffbr); jsonErr == nil {
			return &ffbr // Make sure this is the error that gets persisted now
		}
	}

	return e
}
