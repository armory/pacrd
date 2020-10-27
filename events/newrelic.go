package events

import (
	"fmt"
	"github.com/armory-io/pacrd/api/v1alpha1"
	"github.com/armory/plank"
	"github.com/mitchellh/mapstructure"
	newrelic "github.com/newrelic/go-agent"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)


type NewRelicClient struct {
	Application newrelic.Application
	pacapps		map[string]bool
	apppipes	map[string]bool
}

func (client *NewRelicClient) SendEvent( eventName string, value map[string]interface{} ) {
	if client.Application != nil {
		if !client.IsTimeToSend(){
			return
		}
		// We just need to have the eventName to know reconciliations
		txn := client.Application.StartTransaction(eventName, nil, nil )
		defer txn.End()

		if val, ok := value["TypeMeta"]; ok {
			var typeMeta = metav1.TypeMeta{}
			mapstructure.Decode(val, &typeMeta)
			if typeMeta.Kind == "Pipeline" {
				var pipe = v1alpha1.Pipeline{}
				errdecpipe := mapstructure.Decode(value, &pipe)
				if errdecpipe == nil {
					client.apppipes[pipe.Spec.Application + pipe.Name] = false
					client.Application.RecordCustomMetric("totalPipelines", float64(len(client.apppipes)))
					return
				}
			}
			if typeMeta.Kind == "Application" {
				var app = v1alpha1.Application{}
				errdec  := mapstructure.Decode(value, &app)
				if errdec == nil {
					client.pacapps[app.Name] = false
					client.Application.RecordCustomMetric("totalApps", float64(len(client.pacapps)))
					return
				}
			}
		}
	}
}

func (client *NewRelicClient) SendError(eventName string, trace error) {

	txn := client.Application.StartTransaction(eventName, nil, nil )
	defer txn.End()
	txn.NoticeError(trace)

}

func (client *NewRelicClient) SendPipelineStages( pipeline plank.Pipeline ) {
	if !client.IsTimeToSend(){
		return
	}
	for _, stage := range pipeline.Stages {
		if val, ok := stage["type"]; ok {
			client.SendEvent(fmt.Sprintf("%v", val), stage)
		}
	}
}



func NewNewRelicEventClient(settings EventClientSettings) (EventClient, error) {
	config := newrelic.NewConfig(settings.AppName, settings.ApiKey)
	app, err := newrelic.NewApplication( config )
	// If an application could not be created then err will reveal why.
	if err != nil {
		return nil, err
	}
	return &NewRelicClient{
		Application: app,
		pacapps: make(map[string]bool),
		apppipes: make(map[string]bool),
	}, err
}


func(client *NewRelicClient) IsTimeToSend()  bool{
	// Every hour metrics will be send for 3 minutes
	if time.Now().Minute() <= 2 {
		return true
	}
	return false
}
