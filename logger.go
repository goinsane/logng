package logng

import (
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
	time               time.Time
	prefix             string
	suffix             string
	fields             Fields
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

// Clone clones the Logger.
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
		time:               l.time,
		prefix:             l.prefix,
		suffix:             l.suffix,
		fields:             l.fields.Clone(),
	}
	return l2
}

func (l *Logger) out(severity Severity, message string, err error) {
	if l == nil {
		return
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if !(l.verbose >= l.verbosity) {
		return
	}
	if l.output != nil && l.severity >= severity && l.verbose >= l.verbosity {
		messageLen := len(l.prefix) + len(message) + len(l.suffix)
		log := &Log{
			Message:   make([]byte, 0, messageLen),
			Error:     err,
			Severity:  severity,
			Verbosity: l.verbosity,
			Time:      l.time,
			Fields:    l.fields.Clone(),
		}
		log.Message = append(log.Message, l.prefix...)
		log.Message = append(log.Message, message...)
		log.Message = append(log.Message, l.suffix...)
		if messageLen != 0 && log.Message[messageLen-1] == '\n' {
			log.Message = log.Message[:messageLen-1]
		}
		if log.Time.IsZero() {
			log.Time = time.Now()
		}
		log.StackCaller = NewStackTrace(ProgramCounters(1, 5)).Caller(0)
		if l.stackTraceSeverity >= severity {
			log.StackTrace = NewStackTrace(ProgramCounters(64, 5))
		}
		l.output.Log(log)
	}
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

// Print logs a log which has the Logger's print severity.
func (l *Logger) Print(args ...interface{}) {
	if l == nil {
		return
	}
	l.log(l.printSeverity, args...)
}

// Printf logs a log which has the Logger's print severity.
func (l *Logger) Printf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.logf(l.printSeverity, format, args...)
}

// Println logs a log which has the Logger's print severity.
func (l *Logger) Println(args ...interface{}) {
	if l == nil {
		return
	}
	l.logln(l.printSeverity, args...)
}

// SetOutput sets the Logger's output.
// It returns underlying Logger.
func (l *Logger) SetOutput(output Output) *Logger {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
	return l
}

// SetSeverity sets the Logger's severity.
// If severity is invalid, it sets SeverityInfo.
// It returns underlying Logger.
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

// SetVerbose sets the Logger's verbose.
// It returns underlying Logger.
func (l *Logger) SetVerbose(verbose Verbose) *Logger {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbose = verbose
	return l
}

// SetPrintSeverity sets the Logger's severity level which is using with Print methods.
// If printSeverity is invalid, it sets SeverityInfo.
// It returns underlying Logger.
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

// SetStackTraceSeverity sets the Logger's severity level which saves stack trace into Log.
// If stackTraceSeverity is invalid, it sets SeverityNone.
// It returns underlying Logger.
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

// V clones the Logger if the Logger's verbose is greater or equal to given verbosity, otherwise returns nil.
func (l *Logger) V(verbosity Verbose) *Logger {
	if l == nil {
		return nil
	}
	if !(l.verbose >= verbosity) {
		return nil
	}
	l2 := l.Clone()
	l2.verbosity = verbosity
	return l2
}

// WithTime clones the Logger with given time.
func (l *Logger) WithTime(tm time.Time) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.time = tm
	return l2
}

// WithPrefix clones the Logger and adds the given prefix to end of the underlying prefix.
func (l *Logger) WithPrefix(args ...interface{}) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.prefix += fmt.Sprint(args...)
	return l2
}

// WithPrefixf clones the Logger and adds the given prefix to end of the underlying prefix.
func (l *Logger) WithPrefixf(format string, args ...interface{}) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.prefix += fmt.Sprintf(format, args...)
	return l2
}

// WithSuffix clones the Logger and adds the given suffix to start of the underlying suffix.
func (l *Logger) WithSuffix(args ...interface{}) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.suffix = fmt.Sprint(args...) + l2.suffix
	return l2
}

// WithSuffixf clones the Logger and adds the given suffix to start of the underlying suffix.
func (l *Logger) WithSuffixf(format string, args ...interface{}) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.suffix = fmt.Sprintf(format, args...) + l2.suffix
	return l2
}

// WithFields clones the Logger with given fields.
func (l *Logger) WithFields(fields ...Field) *Logger {
	if l == nil {
		return nil
	}
	l2 := l.Clone()
	l2.fields = append(l2.fields, fields...)
	return l2
}

// WithFieldKeyVals clones the Logger with given key and values of Field.
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

// WithFieldMap clones the Logger with given fieldMap.
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
