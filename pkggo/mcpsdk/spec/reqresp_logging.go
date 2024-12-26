package spec

// A request from the client to the server, to enable or adjust logging.
type SetLevelRequest struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params SetLevelRequestParams `json:"params" yaml:"params" mapstructure:"params"`
}

type SetLevelRequestParams struct {
	// The level of logging that the client wants to receive from the server. The
	// server should send all logs at this level and higher (i.e., more severe) to the
	// client as notifications/logging/message.
	Level LoggingLevel `json:"level" yaml:"level" mapstructure:"level"`
}

// Notification of a log message passed from server to client. If no
// logging/setLevel request has been sent from the client, the server MAY decide
// which messages to send automatically.
type LoggingMessageNotification struct {
	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params LoggingMessageNotificationParams `json:"params" yaml:"params" mapstructure:"params"`
}

type LoggingMessageNotificationParams struct {
	// The data to be logged, such as a string message or an object. Any JSON
	// serializable type is allowed here.
	Data interface{} `json:"data" yaml:"data" mapstructure:"data"`

	// The severity of this log message.
	Level LoggingLevel `json:"level" yaml:"level" mapstructure:"level"`

	// An optional name of the logger issuing this message.
	Logger *string `json:"logger,omitempty" yaml:"logger,omitempty" mapstructure:"logger,omitempty"`
}
