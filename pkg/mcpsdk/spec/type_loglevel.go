package spec

var enumValuesLoggingLevel = []string{
	"alert",
	"critical",
	"debug",
	"emergency",
	"error",
	"info",
	"notice",
	"warning",
}

// LoggingLevel to support.
type LoggingLevel struct {
	*StringUnion
}

// NewLoggingLevel creates a new LoggingLevel with the provided value.
func NewLoggingLevel(value string) *LoggingLevel {
	stringUnion := NewStringUnion(enumValuesLoggingLevel...)
	_ = stringUnion.SetValue(value)
	return &LoggingLevel{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for LoggingLevel.
func (r *LoggingLevel) UnmarshalJSON(b []byte) error {
	if r.StringUnion == nil {
		// Initialize with allowed values if not already initialized.
		r.StringUnion = NewStringUnion(enumValuesLoggingLevel...)
	}
	return r.StringUnion.UnmarshalJSON(b)
}

// MarshalJSON implements json.Marshaler for LoggingLevel.
func (r *LoggingLevel) MarshalJSON() ([]byte, error) {
	return r.StringUnion.MarshalJSON()
}

var (
	LoggingLevelAlert     *LoggingLevel = NewLoggingLevel("alert")
	LoggingLevelCritical  *LoggingLevel = NewLoggingLevel("critical")
	LoggingLevelDebug     *LoggingLevel = NewLoggingLevel("debug")
	LoggingLevelEmergency *LoggingLevel = NewLoggingLevel("emergency")
	LoggingLevelError     *LoggingLevel = NewLoggingLevel("error")
	LoggingLevelInfo      *LoggingLevel = NewLoggingLevel("info")
	LoggingLevelNotice    *LoggingLevel = NewLoggingLevel("notice")
	LoggingLevelWarning   *LoggingLevel = NewLoggingLevel("warning")
)
