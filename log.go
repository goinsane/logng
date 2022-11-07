package logng

import (
	"time"
)

// Log carries the log.
type Log struct {
	Message     []byte
	Error       error
	Severity    Severity
	Verbosity   Verbose
	Time        time.Time
	Fields      Fields
	StackCaller StackCaller
	StackTrace  *StackTrace
}

// Clone duplicates the Log.
func (l *Log) Clone() *Log {
	if l == nil {
		return nil
	}
	l2 := &Log{
		Message:     nil,
		Error:       l.Error,
		Severity:    l.Severity,
		Verbosity:   l.Verbosity,
		Time:        l.Time,
		Fields:      l.Fields.Clone(),
		StackCaller: l.StackCaller,
		StackTrace:  l.StackTrace.Clone(),
	}
	if l.Message != nil {
		l2.Message = make([]byte, len(l.Message))
		copy(l2.Message, l.Message)
	}
	return l2
}
