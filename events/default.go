package events

import "github.com/armory/plank"

type DefaultClient struct {
}

func (client *DefaultClient) SendEvent( eventName string, value map[string]interface{} ) {
}

func (client *DefaultClient) SendError(eventName string, trace error) {
}

func (client *DefaultClient) SendPipelineStages( pipeline plank.Pipeline ) {
}
