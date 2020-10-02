package events

import (
	"fmt"
	"github.com/armory/plank"
	newrelic "github.com/newrelic/go-agent"
)

type NewRelicClient struct {
	Application newrelic.Application
}

func (client *NewRelicClient) SendEvent( eventName string ) {
	if client.Application != nil {

		// We just need to have the eventName to know reconciliations
		txn := client.Application.StartTransaction(eventName, nil, nil )
		defer txn.End()

	}
}

func (client *NewRelicClient) SendError(eventName string, trace error) {
	if client.Application != nil {

		txn := client.Application.StartTransaction(eventName, nil, nil )
		defer txn.End()
		txn.NoticeError(trace)

	}
}

func (client *NewRelicClient) SendPipelineStages( pipeline plank.Pipeline ) {
	if client.Application != nil {
		for _, stage := range pipeline.Stages {
			if val, ok := stage["type"]; ok {
				client.SendEvent(fmt.Sprintf("%v", val))
			}
		}
	}
}



func NewNewRelicEventClient(settings EventClientSettings) (EventClient, error) {
	config := newrelic.NewConfig(settings.AppName, settings.ApiKey)
	app, err := newrelic.NewApplication( config )
	// If an application could not be created then err will reveal why.
	return &NewRelicClient{
		Application: app,
	}, err
}
