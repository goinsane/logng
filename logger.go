package logng

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// Logger provides a logger for leveled and structured logging.
type Logger struct {
	mu                 sync.RWMutex
	output             Output
	severity           Severity
	verbose            Verbose
	printSeverity      Severity
	stackTraceSeverity Severity
	verbosity          Verbose
	time               *time.Time
	prefix             string
	suffix             string
	fields             Fields
	ctxErrVerbosity    Verbose
}

// NewLogger creates a new Logger. If severity is invalid, it sets SeverityInfo.
func NewLogger(output Output, severity Severity, verbose Verbose) *Logger {
	if !severity.IsValid() {
		severity = SeverityInfo
	}
	return &Logger{
		output:             output,
		severity:           severity,
		verbose:            verbose,
		printSeverity:      SeverityInfo,
		stackTraceSeverity: SeverityNone,
	}
}

// Clone clones the underlying Logger.
func (l *Logger) Clone() *Logger {
	if l == nil {
		return nil
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	l2 := &Logger{
		output:             l.output,
		severity:           l.severity,
		verbose:            l.verbose,
		printSeverity:      l.printSeverity,
		stackTraceSeverity: l.stackTraceSeverity,
		verbosity:          l.verbosity,
		time:               nil,
		prefix:             l.prefix,
		suffix:             l.suffix,
		fields:             l.fields.Clone(),
		ctxErrVerbosity:    l.ctxErrVerbosity,
	}
	if l.time != nil {
		tm := *l.time
		l2.time = &tm
	}
	return l2
}

func (l *Logger) out(severity Severity, message string, err error) {
	if l == nil {
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.output == nil {
		return
	}
	if l.severity < severity {
		return
	}
	if l.verbose < l.verbosity {
		return
	}
	if (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) && l.verbose < l.ctxErrVerbosity {
		return
	}

	messageLen := len(l.prefix) + len(message) + len(l.suffix)

	log := &Log{
		Message:   make([]byte, 0, messageLen),
		Error:     err,
		Severity:  severity,
		Verbosity: l.verbosity,
		Fields:    l.fields.Clone(),
	}

	log.Message = append(log.Message, l.prefix...)
	log.Message = append(log.Message, message...)
	log.Message = append(log.Message, l.suffix...)
	if messageLen > 0 && log.Message[messageLen-1] == '\n' {
		log.Message = log.Message[:messageLen-1]
	}

	if l.time != nil {
		log.Time = *l.time
	} else {
		log.Time = time.Now()
	}

	includeStackTrace := l.stackTraceSeverity >= severity

	pcSize := 1
	if includeStackTrace {
		pcSize = 64
	}
	pc := ProgramCounters(pcSize, 5)
	st := NewStackTrace(pc)

	if st.SizeOfCallers() > 0 {
		log.StackCaller = st.Caller(0)
	}

	if includeStackTrace {
		log.StackTrace = st
	}

	l.output.Log(log)
}

func (l *Logger) log(severity Severity, args ...interface{}) {
	var err error
	for _, arg := range args {
		if e, ok := arg.(error); ok {
			err = e
			break
		}
	}
	l.out(severity, fmt.Sprint(args...), err)
}

func (l *Logger) logf(severity Severity, format string, args ...interface{}) {
	var err error
	wErr := fmt.Errorf(format, args...)
	if e, ok := wErr.(wrappedError); ok {
		err = e.Unwrap()
	}
	l.out(severity, wErr.Error(), err)
}

func (l *Logger) logln(severity Severity, args ...interface{}) {
	var err error
	for _, arg := range args {
		if e, ok := arg.(error); ok {
			err = e
			break
		}
	}
	l.out(severity, fmt.Sprintln(args...), err)
}

// Fatal logs to the FATAL severity logs, then calls os.Exit(1).
func (l *Logger) Fatal(args ...interface{}) {
	l.log(SeverityFatal, args...)
	os.Exit(1)
}

// Fatalf logs to the FATAL severity logs, then calls os.Exit(1).
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logf(SeverityFatal, format, args...)
	os.Exit(1)
}

// Fatalln logs to the FATAL severity logs, then calls os.Exit(1).
func (l *Logger) Fatalln(args ...interface{}) {
	l.logln(SeverityFatal, args...)
	os.Exit(1)
}

// Error logs to the ERROR severity logs.
func (l *Logger) Error(args ...interface{}) {
	l.log(SeverityError, args...)
}

// Errorf logs to the ERROR severity logs.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(SeverityError, format, args...)
}

// Errorln logs to the ERROR severity logs.
func (l *Logger) Errorln(args ...interface{}) {
	l.logln(SeverityError, args...)
}

// Warning logs to the WARNING severity logs.
func (l *Logger) Warning(args ...interface{}) {
	l.log(SeverityWarning, args...)
}

// Warningf logs to the WARNING severity logs.
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.logf(SeverityWarning, format, args...)
}

// Warningln logs to the WARNING severity logs.
func (l *Logger) Warningln(args ...interface{}) {
	l.logln(SeverityWarning, args...)
}

// Info logs to the INFO severity logs.
func (l *Logger) Info(args ...interface{}) {
	l.log(SeverityInfo, args...)
}

// Infof logs to the INFO severity logs.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(SeverityInfo, format, args...)
}

// Infoln logs to the INFO severity logs.
func (l *Logger) Infoln(args ...interface{}) {
	l.logln(SeverityInfo, args...)
}

// Debug logs to the DEBUG severity logs.
func (l *Logger) Debug(args ...interface{}) {
	l.log(SeverityDebug, args...)
}

// Debugf logs to the DEBUG severity logs.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(SeverityDebug, format, args...)
}

// Debugln logs to the DEBUG severity logs.
func (l *Logger) Debugln(args ...interface{}) {
	l.logln(SeverityDebug, args...)
}

// Print logs a log which has the underlying Logger's print severity.
func (l *Logger) Print(args ...interface{}) {
	if l == nil {
		return
	}
	l.log(l.printSeverity, args...)
}

// Printf logs a log which has the underlying Logger's print severity.
func (l *Logger) Printf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.logf(l.printSeverity, format, args...)
}

// Println logs a log which has the underlying Logger's print severity.
func (l *Logger) Println(args ...interface{}) {
	if l == nil {
		return
	}
	l.logln(l.printSeverity, args...)
}

// SetOutput sets the underlying Logger's output.
// It returns the underlying Logger.
func (l *Logger) SetOutput(output Output) *Logger {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
	return l
}

// SetSeverity sets the underlying Logger's severity.
// If severity is invalid, it sets SeverityInfo.
// It returns the underlying Logger.
func (l *Logger) SetSeverity(severity Severity) *Logger {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !severity.IsValid() {
		severity = SeverityInfo
	}
	l.severity = severity
	return l
}

// SetVerbose sets the underlying Logger's verbose.
// It returns the underlying Logger.
func (l *Logger) SetVerbose(verbose Verbose) *Logger {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbose = verbose
	return l
}

// SetPrintSeverity sets the underlying Logger's severity level which is using with Print methods.
// If printSeverity is invalid, it sets SeverityInfo.
// It returns the underlying Logger.
// By default, SeverityInfo.
func (l *Logger) SetPrintSeverity(printSeverity Severity) *Logger {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !printSeverity.IsValid() {
		printSeverity = SeverityInfo
	}
	l.printSeverity = printSeverity
	return l
}

// SetStackTraceSeverity sets the underlying Logger's severity level which saves stack trace into Log.
// If stackTraceSeverity is invalid, it sets SeverityNone.
// It returns the underlying Logger.
// By default, SeverityNone.
func (l *Logger) SetStackTraceSeverity(stackTraceSeverity Severity) *Logger {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !stackTraceSeverity.IsValid() {
		stackTraceSeverity = SeverityNone
	}
	l.stackTraceSeverity = stackTraceSeverity
	return l
}

// V clones the underlying Logger with the given verbosity if the underlying Logger's verbose is greater or equal to the given verbosity, otherwise returns nil.
func (l *Logger) V(verbosity Verbose) *Logger {
	if l == nil {
		return nil
	}
	l.mu.RLock()
	if l.verbose < verbosity {
		l.mu.RUnlock()
		return nil
	}
	l.mu.RUnlock()
	return l.WithVerbosity(verbosity)
}

// WithVerbosity clones the underlying Logger with the given verbosity.
func (l *Logger) WithVerbosity(verbosity Verbose) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.verbosity = verbosity
	return l2
}

// WithTime clones the underlying Logger with the given time.
func (l *Logger) WithTime(tm time.Time) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.time = &tm
	return l2
}

// WithoutTime clones the underlying Logger without time.
func (l *Logger) WithoutTime() *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.time = nil
	return l2
}

// WithPrefix clones the underlying Logger and adds the given prefix to the end of the underlying prefix.
func (l *Logger) WithPrefix(args ...interface{}) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.prefix += fmt.Sprint(args...)
	return l2
}

// WithPrefixf clones the underlying Logger and adds the given prefix to the end of the underlying prefix.
func (l *Logger) WithPrefixf(format string, args ...interface{}) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.prefix += fmt.Sprintf(format, args...)
	return l2
}

// WithSuffix clones the underlying Logger and adds the given suffix to the beginning of the underlying suffix.
func (l *Logger) WithSuffix(args ...interface{}) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.suffix = fmt.Sprint(args...) + l2.suffix
	return l2
}

// WithSuffixf clones the underlying Logger and adds the given suffix to the beginning of the underlying suffix.
func (l *Logger) WithSuffixf(format string, args ...interface{}) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.suffix = fmt.Sprintf(format, args...) + l2.suffix
	return l2
}

// WithFields clones the underlying Logger with given fields.
func (l *Logger) WithFields(fields ...Field) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.fields = append(l2.fields, fields...)
	return l2
}

// WithFieldKeyVals clones the underlying Logger with given keys and values of Field.
func (l *Logger) WithFieldKeyVals(kvs ...interface{}) *Logger {
	if l == nil {
		return nil
	}
	n := len(kvs) / 2
	fields := make(Fields, 0, n)
	for i := 0; i < n; i++ {
		j := i * 2
		k, v := fmt.Sprintf("%v", kvs[j]), kvs[j+1]
		fields = append(fields, Field{Key: k, Value: v})
	}
	return l.WithFields(fields...)
}

// WithFieldMap clones the underlying Logger with the given fieldMap.
func (l *Logger) WithFieldMap(fieldMap map[string]interface{}) *Logger {
	if l == nil {
		return nil
	}
	fields := make(Fields, 0, len(fieldMap))
	for k, v := range fieldMap {
		fields = append(fields, Field{Key: k, Value: v})
	}
	return l.WithFields(fields...)
}

// WithCtxErrVerbosity clones the underlying Logger with context error verbosity.
// If the log has an error and the error is an context error, the given value is used as verbosity.
func (l *Logger) WithCtxErrVerbosity(verbosity Verbose) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.ctxErrVerbosity = verbosity
	return l2
}
