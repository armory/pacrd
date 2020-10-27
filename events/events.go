package events

import (
	"github.com/armory/plank/v3"
)

type EventClient interface {
	SendEvent(string, map[string]interface{})
	SendError(string, error)
	SendPipelineStages(plank.Pipeline)
}

type EventClientSettings struct {
	AppName string
	ApiKey 	string
}