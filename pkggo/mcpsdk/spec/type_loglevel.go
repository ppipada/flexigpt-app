package spec

import "github.com/flexigpt/flexiui/pkggo/mcpsdk/spec/customtype"

var enumValues_LoggingLevel = []string{
	"alert",
	"critical",
	"debug",
	"emergency",
	"error",
	"info",
	"notice",
	"warning",
}

// LoggingLevel
type LoggingLevel struct {
	*customtype.StringUnion
}

// NewLoggingLevel creates a new LoggingLevel with the provided value.
func NewLoggingLevel(value string) *LoggingLevel {
	stringUnion := customtype.NewStringUnion(enumValues_LoggingLevel...)
	_ = stringUnion.SetValue(value)
	return &LoggingLevel{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for LoggingLevel.
func (r *LoggingLevel) UnmarshalJSON(b []byte) error {
	if r.StringUnion == nil {
		// Initialize with allowed values if not already initialized
		r.StringUnion = customtype.NewStringUnion(enumValues_LoggingLevel...)
	}
	return r.StringUnion.UnmarshalJSON(b)
}

// MarshalJSON implements json.Marshaler for LoggingLevel.
func (r *LoggingLevel) MarshalJSON() ([]byte, error) {
	return r.StringUnion.MarshalJSON()
}

var LoggingLevelAlert *LoggingLevel = NewLoggingLevel("alert")
var LoggingLevelCritical *LoggingLevel = NewLoggingLevel("critical")
var LoggingLevelDebug *LoggingLevel = NewLoggingLevel("debug")
var LoggingLevelEmergency *LoggingLevel = NewLoggingLevel("emergency")
var LoggingLevelError *LoggingLevel = NewLoggingLevel("error")
var LoggingLevelInfo *LoggingLevel = NewLoggingLevel("info")
var LoggingLevelNotice *LoggingLevel = NewLoggingLevel("notice")
var LoggingLevelWarning *LoggingLevel = NewLoggingLevel("warning")
