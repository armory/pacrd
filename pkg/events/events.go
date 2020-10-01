package events

type EventClient interface {
	SendEvent(string, map[string]interface{})
	SendError(string, error)
}

type EventClientSettings struct {
	AppName string
	ApiKey 	string
}