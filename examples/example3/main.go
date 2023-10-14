package main

import (
	"os"
	"time"

	"github.com/goinsane/logng"
)

var (
	testTime, _ = time.ParseInLocation("2006-01-02T15:04:05", "2010-11-12T13:14:15", time.Local)
)

func main() {
	// reset logng for previous changes if it is running in go test.
	logng.Reset()

	// set JSONOutput2.
	output := logng.NewJSONOutput2(os.Stdout, logng.JSONOutput2FlagDefault)
	logng.SetOutput(output)

	// log by Severity.
	// default severity is SeverityInfo.
	// default verbose is 0.
	logng.Debug("this is debug log. but it won't be shown.")
	logng.Info("this is info log.")
	logng.Warning("this is warning log.")
	logng.V(1).Error("this is error log, verbosity 1. but it won't be shown.")

	// SetSeverity()
	// default severity is SeverityInfo.
	logng.SetSeverity(logng.SeverityDebug)
	logng.Debug("this is debug log. it will now be shown.")

	// SetVerbose() and V()
	// default verbose is 0
	logng.SetVerbose(1)
	logng.V(1).Error("this is error log, verbosity 1. it will now be shown.")
	logng.V(2).Warning("this is warning log, verbosity 2. it won't be shown.")

	// SetPrintSeverity()
	// default print severity is SeverityInfo.
	logng.Print("this log will be shown as info log.")
	logng.SetPrintSeverity(logng.SeverityWarning)
	logng.Print("this log will now be shown as warning log.")

	// SetStackTraceSeverity()
	// default stack trace severity is none.
	logng.SetStackTraceSeverity(logng.SeverityWarning)
	logng.Warning("this is warning log. you can see stack trace for this log.")
	logng.Error("this is error log. you can still see stack trace for this log.")
	logng.Info("this is info log. stack trace won't be shown for this log.")

	// WithTime()
	logng.WithTime(testTime).Info("this is info log with custom time.")

	// WithFieldKeyVals()
	logng.WithFieldKeyVals("key1", "val1", "key2", "val2", "key3", "val3", "key1", "val1-2", "key2", "val2-2").Info("this is info log with several fields.")

	output.SetFlags(logng.JSONOutput2FlagSeverity |
		logng.JSONOutput2FlagTime |
		logng.JSONOutput2FlagTimestamp |
		logng.JSONOutput2FlagUTC |
		logng.JSONOutput2FlagSeverityLevel |
		logng.JSONOutput2FlagVerbosity |
		logng.JSONOutput2FlagShortFunc |
		logng.JSONOutput2FlagLongFile)
	logng.Info("test new flags")
}
