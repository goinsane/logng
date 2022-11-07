package logng

import (
	"strings"
)

// Severity describes severity level of Log.
type Severity int

const (
	// SeverityNone is none or unspecified severity level
	SeverityNone Severity = iota

	// SeverityFatal is fatal severity level
	SeverityFatal

	// SeverityError is error severity level
	SeverityError

	// SeverityWarning is warning severity level
	SeverityWarning

	// SeverityInfo is info severity level
	SeverityInfo

	// SeverityDebug is debug severity level
	SeverityDebug
)

// IsValid returns whether s is valid.
func (s Severity) IsValid() bool {
	return s.CheckValid() == nil
}

// CheckValid returns ErrInvalidSeverity for invalid s.
func (s Severity) CheckValid() error {
	if !(SeverityNone <= s && s <= SeverityDebug) {
		return ErrInvalidSeverity
	}
	return nil
}

// String is implementation of fmt.Stringer.
func (s Severity) String() string {
	text, _ := s.MarshalText()
	return string(text)
}

// MarshalText is implementation of encoding.TextMarshaler.
// If s is invalid, it returns nil and result of Severity.CheckValid.
func (s Severity) MarshalText() (text []byte, err error) {
	if e := s.CheckValid(); e != nil {
		return nil, e
	}
	var str string
	switch s {
	case SeverityNone:
		str = "NONE"
	case SeverityFatal:
		str = "FATAL"
	case SeverityError:
		str = "ERROR"
	case SeverityWarning:
		str = "WARNING"
	case SeverityInfo:
		str = "INFO"
	case SeverityDebug:
		str = "DEBUG"
	default:
		panic("invalid severity")
	}
	return []byte(str), nil
}

// UnmarshalText is implementation of encoding.UnmarshalText.
// If text is unknown, it returns ErrUnknownSeverity.
func (s *Severity) UnmarshalText(text []byte) error {
	switch str := strings.ToUpper(string(text)); str {
	case "NONE":
		*s = SeverityNone
	case "FATAL":
		*s = SeverityFatal
	case "ERROR":
		*s = SeverityError
	case "WARNING":
		*s = SeverityWarning
	case "INFO":
		*s = SeverityInfo
	case "DEBUG":
		*s = SeverityDebug
	default:
		return ErrUnknownSeverity
	}
	return nil
}
