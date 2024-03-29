// Package logng provides structured and leveled logging.
package logng

import (
	"io"
	"os"
	"time"
)

// Reset resets the default Logger and the default TextOutput.
func Reset() {
	SetOutput(defaultTextOutput)
	SetSeverity(SeverityInfo)
	SetVerbose(0)
	SetPrintSeverity(SeverityInfo)
	SetStackTraceSeverity(SeverityNone)
	SetStackTraceSize(64)
	SetTextOutputWriter(defaultTextOutputWriter)
	SetTextOutputFlags(TextOutputFlagDefault)
}

var (
	defaultLogger = NewLogger(defaultTextOutput, SeverityInfo, 0)
)

// DefaultLogger returns the default Logger.
func DefaultLogger() *Logger {
	return defaultLogger
}

// Clone clones the default Logger.
func Clone() *Logger {
	return defaultLogger.Clone()
}

// Fatal logs to the FATAL severity logs to the default Logger, then calls os.Exit(1).
func Fatal(args ...interface{}) {
	defaultLogger.log(SeverityFatal, args...)
	os.Exit(1)
}

// Fatalf logs to the FATAL severity logs to the default Logger, then calls os.Exit(1).
func Fatalf(format string, args ...interface{}) {
	defaultLogger.logf(SeverityFatal, format, args...)
	os.Exit(1)
}

// Fatalln logs to the FATAL severity logs to the default Logger, then calls os.Exit(1).
func Fatalln(args ...interface{}) {
	defaultLogger.logln(SeverityFatal, args...)
	os.Exit(1)
}

// Error logs to the ERROR severity logs to the default Logger.
func Error(args ...interface{}) {
	defaultLogger.log(SeverityError, args...)
}

// Errorf logs to the ERROR severity logs to the default Logger.
func Errorf(format string, args ...interface{}) {
	defaultLogger.logf(SeverityError, format, args...)
}

// Errorln logs to the ERROR severity logs to the default Logger.
func Errorln(args ...interface{}) {
	defaultLogger.logln(SeverityError, args...)
}

// Warning logs to the WARNING severity logs to the default Logger.
func Warning(args ...interface{}) {
	defaultLogger.log(SeverityWarning, args...)
}

// Warningf logs to the WARNING severity logs to the default Logger.
func Warningf(format string, args ...interface{}) {
	defaultLogger.logf(SeverityWarning, format, args...)
}

// Warningln logs to the WARNING severity logs to the default Logger.
func Warningln(args ...interface{}) {
	defaultLogger.logln(SeverityWarning, args...)
}

// Info logs to the INFO severity logs to the default Logger.
func Info(args ...interface{}) {
	defaultLogger.log(SeverityInfo, args...)
}

// Infof logs to the INFO severity logs to the default Logger.
func Infof(format string, args ...interface{}) {
	defaultLogger.logf(SeverityInfo, format, args...)
}

// Infoln logs to the INFO severity logs to the default Logger.
func Infoln(args ...interface{}) {
	defaultLogger.logln(SeverityInfo, args...)
}

// Debug logs to the DEBUG severity logs to the default Logger.
func Debug(args ...interface{}) {
	defaultLogger.log(SeverityDebug, args...)
}

// Debugf logs to the DEBUG severity logs to the default Logger.
func Debugf(format string, args ...interface{}) {
	defaultLogger.logf(SeverityDebug, format, args...)
}

// Debugln logs to the DEBUG severity logs to the default Logger.
func Debugln(args ...interface{}) {
	defaultLogger.logln(SeverityDebug, args...)
}

// Print logs a log which has the default Logger's print severity to the default Logger.
func Print(args ...interface{}) {
	defaultLogger.log(severityPrint, args...)
}

// Printf logs a log which has the default Logger's print severity to the default Logger.
func Printf(format string, args ...interface{}) {
	defaultLogger.logf(severityPrint, format, args...)
}

// Println logs a log which has the default Logger's print severity to the default Logger.
func Println(args ...interface{}) {
	defaultLogger.logln(severityPrint, args...)
}

// SetOutput sets the default Logger's output.
// It returns the default Logger.
// By default, the default TextOutput.
func SetOutput(output Output) *Logger {
	return defaultLogger.SetOutput(output)
}

// SetSeverity sets the default Logger's severity.
// If severity is invalid, it sets SeverityInfo.
// It returns the default Logger.
// By default, SeverityInfo.
func SetSeverity(severity Severity) *Logger {
	return defaultLogger.SetSeverity(severity)
}

// SetVerbose sets the default Logger's verbose.
// It returns the default Logger.
// By default, 0.
func SetVerbose(verbose Verbose) *Logger {
	return defaultLogger.SetVerbose(verbose)
}

// SetPrintSeverity sets the default Logger's severity level which is using with Print methods.
// If printSeverity is invalid, it sets SeverityInfo.
// It returns the default Logger.
// By default, SeverityInfo.
func SetPrintSeverity(printSeverity Severity) *Logger {
	return defaultLogger.SetPrintSeverity(printSeverity)
}

// SetStackTraceSeverity sets the default Logger's severity level which saves stack trace into Log.
// If stackTraceSeverity is invalid, it sets SeverityNone.
// It returns the default Logger.
// By default, SeverityNone.
func SetStackTraceSeverity(stackTraceSeverity Severity) *Logger {
	return defaultLogger.SetStackTraceSeverity(stackTraceSeverity)
}

// SetStackTraceSize sets the maximum program counter size of the stack trace for the default Logger.
// If stackTraceSize is out of range, it sets 64. The range is 1 to 16384 each included.
// It returns the default Logger.
// By default, 64.
func SetStackTraceSize(stackTraceSize int) *Logger {
	return defaultLogger.SetStackTraceSize(stackTraceSize)
}

// V clones the default Logger with the given verbosity if the default Logger's verbose is greater or equal to the given verbosity, otherwise returns nil.
func V(verbosity Verbose) *Logger {
	return defaultLogger.V(verbosity)
}

// WithVerbosity clones the default Logger with the given verbosity.
func WithVerbosity(verbosity Verbose) *Logger {
	return defaultLogger.WithVerbosity(verbosity)
}

// WithTime clones the default Logger with the given time.
func WithTime(tm time.Time) *Logger {
	return defaultLogger.WithTime(tm)
}

// WithoutTime clones the default Logger without time.
func WithoutTime() *Logger {
	return defaultLogger.WithoutTime()
}

// WithPrefix clones the default Logger and adds the given prefix to the end of the underlying prefix.
func WithPrefix(args ...interface{}) *Logger {
	return defaultLogger.WithPrefix(args...)
}

// WithPrefixf clones the default Logger and adds the given prefix to the end of the underlying prefix.
func WithPrefixf(format string, args ...interface{}) *Logger {
	return defaultLogger.WithPrefixf(format, args...)
}

// WithSuffix clones the default Logger and adds the given suffix to the beginning of the underlying suffix.
func WithSuffix(args ...interface{}) *Logger {
	return defaultLogger.WithSuffix(args...)
}

// WithSuffixf clones the default Logger and adds the given suffix to the beginning of the underlying suffix.
func WithSuffixf(format string, args ...interface{}) *Logger {
	return defaultLogger.WithSuffixf(format, args...)
}

// WithFields clones the default Logger with given fields.
func WithFields(fields ...Field) *Logger {
	return defaultLogger.WithFields(fields...)
}

// WithFieldKeyVals clones the default Logger with given keys and values of Field.
func WithFieldKeyVals(kvs ...interface{}) *Logger {
	return defaultLogger.WithFieldKeyVals(kvs...)
}

// WithFieldMap clones the default Logger with the given field map.
func WithFieldMap(fieldMap map[string]interface{}) *Logger {
	return defaultLogger.WithFieldMap(fieldMap)
}

// WithCtxErrVerbosity clones the default Logger with context error verbosity.
// If the log has an error and the error is an context error, the given value is used as verbosity.
func WithCtxErrVerbosity(verbosity Verbose) *Logger {
	return defaultLogger.WithCtxErrVerbosity(verbosity)
}

var (
	defaultTextOutput       = NewTextOutput(defaultTextOutputWriter, TextOutputFlagDefault)
	defaultTextOutputWriter = os.Stderr
)

// DefaultTextOutput returns the default TextOutput.
func DefaultTextOutput() *TextOutput {
	return defaultTextOutput
}

// SetTextOutputWriter sets the default TextOutput's writer.
// It returns the default TextOutput.
// By default, os.Stderr.
func SetTextOutputWriter(w io.Writer) *TextOutput {
	return defaultTextOutput.SetWriter(w)
}

// SetTextOutputFlags sets the default TextOutput's flags.
// It returns the default TextOutput.
// By default, TextOutputFlagDefault.
func SetTextOutputFlags(flags TextOutputFlag) *TextOutput {
	return defaultTextOutput.SetFlags(flags)
}
