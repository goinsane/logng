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
	// change writer of default output to stdout from stderr.
	logng.SetTextOutputWriter(os.Stdout)

	// log by Severity.
	// default severity is SeverityInfo.
	// default verbose is 0.
	logng.Debug("this is debug log. but it won't be shown.")
	logng.Info("this is info log.")
	logng.Warning("this is warning log.")
	logng.V(1).Error("this is error log, verbosity 1. but it won't be shown.")

	// SetSeverity()
	logng.SetSeverity(logng.SeverityDebug)
	logng.Debug("this is debug log. it will now be shown.")

	// SetVerbose() and V()
	logng.SetVerbose(1)
	logng.V(1).Error("this is error log, verbosity 1. it will now be shown.")
	logng.V(2).Warning("this is warning log, verbosity 2. it won't be shown.")

	// SetFlags()
	// default flags is TextOutputFlagDefault.
	logng.SetTextOutputFlags(logng.TextOutputFlagDefault | logng.TextOutputFlagShortFile)
	logng.Info("this is info log. you can see file name and line in this log.")

	// log using Print.
	// default print severity is SeverityInfo.
	logng.Print("this log will be shown as info log.")

	// SetPrintSeverity()
	logng.SetPrintSeverity(logng.SeverityWarning)
	logng.Print("this log will now be shown as warning log.")

	// SetStackTraceSeverity()
	// default stack trace severity is none.
	logng.SetStackTraceSeverity(logng.SeverityWarning)
	logng.Warning("this is warning log. you can see stack trace end of this log.")
	logng.Error("this is error log. you can still see stack trace end of this log.")
	logng.Info("this is info log. stack trace won't be shown end of this log.")

	// WithTime()
	logng.WithTime(testTime).Info("this is info log with custom time.")

	// WithFieldKeyVals()
	logng.WithFieldKeyVals("key1", "val1", "key2", "val2", "key3", "val3", "key1", "val1-2", "key2", "val2-2").Info("this is info log with several fields.")

	// multi-line logs
	logng.Info("this is\nmulti-line log with file name")
	logng.SetTextOutputFlags(logng.TextOutputFlagDefault)
	logng.Info("this is\nmulti-line log")
	logng.WithFieldKeyVals("key1", "val1").Info("this is\nmulti-line log with key vals")
}
