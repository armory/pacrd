package events

import (
	"github.com/armory/plank"
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