package events

import (
	"encoding/json"
	newrelic "github.com/newrelic/go-agent"
)

type NewRelicClient struct {
	Application newrelic.Application
}

func (client *NewRelicClient) SendEvent( eventName string ,eventValue map[string]interface{}) {
	if client.Application != nil {

		jsonBytes, _ := json.Marshal(eventValue)
		txn := client.Application.StartTransaction(eventName, nil, nil )
		defer txn.End()
		txn.SetName(eventName)
		txn.AddAttribute("content", string(jsonBytes) )

	}
}

func (client *NewRelicClient) SendError(eventName string, trace error) {
	if client.Application != nil {

		txn := client.Application.StartTransaction(eventName, nil, nil )
		defer txn.End()
		txn.NoticeError(trace)

	}
}


func NewNewRelicEventClient(settings EventClientSettings) (EventClient, error) {
	app, err := newrelic.NewApplication( newrelic.NewConfig(settings.AppName, settings.ApiKey));
	// If an application could not be created then err will reveal why.
	return &NewRelicClient{
		Application: app,
	}, err
}
