package logng_test

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/goinsane/logng/v2"
)

func Example() {
	// reset logng for previous changes.
	logng.Reset()
	// change writer of default output to stdout from stderr.
	logng.SetTextOutputWriter(os.Stdout)

	// log by severity and verbosity.
	// default Logger's severity is SeverityInfo.
	// default Logger's verbose is 0.
	logng.Debug("this is debug log. it won't be shown.")
	logng.Info("this is info log.")
	logng.Warning("this is warning log.")
	logng.V(1).Error("this is error log, verbosity 1. it won't be shown.")

	// SetSeverity()
	logng.SetSeverity(logng.SeverityDebug)
	logng.Debug("this is debug log. it will now be shown.")

	// SetVerbose() and V()
	logng.SetVerbose(1)
	logng.V(1).Error("this is error log, verbosity 1. it will now be shown.")
	logng.V(2).Warning("this is warning log, verbosity 2. it won't be shown.")

	// SetTextOutputFlags()
	// default flags is FlagDefault.
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

	// WithPrefix()
	logng.WithPrefix("prefix1").Warning("this is warning log with prefix 'prefix1'.")
	logng.WithPrefix("prefix1").WithPrefix("prefix2").Error("this is error log with both of prefixes 'prefix1' and 'prefix2'.")

	// WithFieldKeyVals()
	logng.WithFieldKeyVals("key1", "val1", "key2", "val2", "key3", "val3", "key1", "val1-2", "key2", "val2-2").Info("this is info log with several fields.")

	// SetTextOutputFlags()
	// default flags is TextOutputFlagDefault.
	logng.SetTextOutputFlags(logng.TextOutputFlagDefault | logng.TextOutputFlagShortFile)
	logng.Info("this is info log. you can see file name and line in this log.")

	// multi-line logs
	logng.Info("this is\nmulti-line log with file name")
	logng.Info("this is\nmulti-line log")
	logng.WithFieldKeyVals("key1", "val1").Info("this is\nmulti-line log with key vals")
}

func Example_test1() {
	// reset logng for previous changes.
	logng.Reset()
	// change writer of default output to stdout from stderr.
	logng.SetTextOutputWriter(os.Stdout)
	// just show severity.
	logng.SetTextOutputFlags(logng.TextOutputFlagSeverity)

	logng.Debug("this is debug log, verbosity 0. it will not be shown.")
	logng.Info("this is info log, verbosity 0.")
	logng.Warning("this is warning log, verbosity 0.")
	logng.Error("this is error log, verbosity 0.")
	logng.Print("this is info log, verbosity 0 caused by Print().")
	logng.V(1).Info("this is info log, verbosity 1. it will not be shown.")

	logng.SetSeverity(logng.SeverityDebug)
	logng.Debug("this is debug log, verbosity 0.")

	logng.SetVerbose(1)
	logng.V(0).Info("this is info log, verbosity 0.")
	logng.V(1).Info("this is info log, verbosity 1.")
	logng.V(2).Info("this is info log, verbosity 2. it will not be shown.")

	logng.SetPrintSeverity(logng.SeverityWarning)
	logng.Print("this is warning log, verbosity 0 caused by Print().")

	logng.Warning("this is warning log, verbosity 0.\nwithout padding.")
	logng.SetTextOutputFlags(logng.TextOutputFlagSeverity | logng.TextOutputFlagPadding)
	logng.Warning("this is warning log, verbosity 0.\nwith padding.")

	logng.SetTextOutputFlags(logng.TextOutputFlagDefault)
	tm, _ := time.ParseInLocation("2006-01-02T15:04:05", "2019-11-13T21:56:24", time.Local)
	logng.WithTime(tm).Info("this is info log, verbosity 0.")

	// Output:
	// INFO - this is info log, verbosity 0.
	// WARNING - this is warning log, verbosity 0.
	// ERROR - this is error log, verbosity 0.
	// INFO - this is info log, verbosity 0 caused by Print().
	// DEBUG - this is debug log, verbosity 0.
	// INFO - this is info log, verbosity 0.
	// INFO - this is info log, verbosity 1.
	// WARNING - this is warning log, verbosity 0 caused by Print().
	// WARNING - this is warning log, verbosity 0.
	// without padding.
	// WARNING - this is warning log, verbosity 0.
	//           with padding.
	// 2019/11/13 21:56:24 INFO - this is info log, verbosity 0.
}

func ExampleSetSeverity() {
	// set logng for this example.
	logng.Reset()
	logng.SetTextOutputWriter(os.Stdout)
	logng.SetTextOutputFlags(logng.TextOutputFlagSeverity)

	logng.SetSeverity(logng.SeverityWarning)
	logng.Debug("this is debug log.")
	logng.Info("this is info log.")
	logng.Warning("this is warning log.")
	logng.Error("this is error log.")

	// Output:
	// WARNING - this is warning log.
	// ERROR - this is error log.
}

func ExampleSetVerbose() {
	// set logng for this example.
	logng.Reset()
	logng.SetTextOutputWriter(os.Stdout)
	logng.SetTextOutputFlags(logng.TextOutputFlagSeverity)

	logng.SetVerbose(1)
	logng.V(0).Info("this is info log, verbosity 0.")
	logng.V(1).Info("this is info log, verbosity 1.")
	logng.V(2).Info("this is info log, verbosity 2. it won't be shown.")
	logng.V(3).Info("this is info log, verbosity 3. it won't be shown.")

	// Output:
	// INFO - this is info log, verbosity 0.
	// INFO - this is info log, verbosity 1.
}

func ExampleSetPrintSeverity() {
	// set logng for this example.
	logng.Reset()
	logng.SetTextOutputWriter(os.Stdout)
	logng.SetTextOutputFlags(logng.TextOutputFlagSeverity)

	logng.SetPrintSeverity(logng.SeverityWarning)
	logng.Print("this is the log.")

	// Output:
	// WARNING - this is the log.
}

func ExampleWithTime() {
	// set logng for this example.
	logng.Reset()
	logng.SetTextOutputWriter(os.Stdout)
	logng.SetTextOutputFlags(logng.TextOutputFlagDefault)

	logng.WithTime(testTime).Info("this is info log with the given time.")

	// Output:
	// 2010/11/12 13:14:15 INFO - this is info log with the given time.
}

func ExampleWithPrefix() {
	// set logng for this example.
	logng.Reset()
	logng.SetTextOutputWriter(os.Stdout)
	logng.SetTextOutputFlags(logng.TextOutputFlagSeverity)

	logng.WithPrefix("prefix1: ").WithPrefix("prefix2: ").Info("this is info log.")

	// Output:
	// INFO - prefix1: prefix2: this is info log.
}

func ExampleWithSuffix() {
	// set logng for this example.
	logng.Reset()
	logng.SetTextOutputWriter(os.Stdout)
	logng.SetTextOutputFlags(logng.TextOutputFlagSeverity)

	logng.WithSuffix(" :suffix1").WithSuffix(" :suffix2").Info("this is info log.")

	// Output:
	// INFO - this is info log. :suffix2 :suffix1
}

func ExampleLogger() {
	logger := logng.NewLogger(logng.NewTextOutput(os.Stdout, logng.TextOutputFlagSeverity),
		logng.SeverityInfo, 2)

	logger.Debug("this is debug log, verbosity 0. it won't be shown.")
	logger.Info("this is info log, verbosity 0.")
	logger.V(0).Debug("this is debug log, verbosity 0. it won't be shown.")
	logger.V(1).Info("this is info log, verbosity 1.")
	logger.V(2).Warning("this is warning log, verbosity 2.")
	logger.V(3).Error("this is error log, verbosity 3. it won't be shown.")

	// Output:
	// INFO - this is info log, verbosity 0.
	// INFO - this is info log, verbosity 1.
	// WARNING - this is warning log, verbosity 2.
}

func BenchmarkInfo(b *testing.B) {
	logng.Reset()
	logng.SetTextOutputWriter(io.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.Info("benchmark")
	}
}

func BenchmarkInfof(b *testing.B) {
	logng.Reset()
	logng.SetTextOutputWriter(io.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.Infof("%s", "benchmark")
	}
}

func BenchmarkInfoln(b *testing.B) {
	logng.Reset()
	logng.SetTextOutputWriter(io.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.Infoln("benchmark")
	}
}

func BenchmarkInfo_withStackTrace(b *testing.B) {
	logng.Reset()
	logng.SetTextOutputWriter(io.Discard)
	logng.SetStackTraceSeverity(logng.SeverityInfo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.Info("benchmark")
	}
}

func BenchmarkInfo_withTextOutputFlagLongFunc(b *testing.B) {
	logng.Reset()
	logng.SetTextOutputWriter(io.Discard)
	logng.SetTextOutputFlags(logng.TextOutputFlagDefault | logng.TextOutputFlagLongFunc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.Info("benchmark")
	}
}

func BenchmarkInfo_withTextOutputFlagShortFunc(b *testing.B) {
	logng.Reset()
	logng.SetTextOutputWriter(io.Discard)
	logng.SetTextOutputFlags(logng.TextOutputFlagDefault | logng.TextOutputFlagShortFunc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.Info("benchmark")
	}
}

func BenchmarkInfo_withFlagLongFile(b *testing.B) {
	logng.Reset()
	logng.SetTextOutputWriter(io.Discard)
	logng.SetTextOutputFlags(logng.TextOutputFlagDefault | logng.TextOutputFlagLongFile)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.Info("benchmark")
	}
}

func BenchmarkInfo_withTextOutputFlagShortFile(b *testing.B) {
	logng.Reset()
	logng.SetTextOutputWriter(io.Discard)
	logng.SetTextOutputFlags(logng.TextOutputFlagDefault | logng.TextOutputFlagShortFile)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.Info("benchmark")
	}
}

func BenchmarkV(b *testing.B) {
	logng.Reset()
	logng.SetVerbose(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.V(1)
	}
}

func BenchmarkWithTime(b *testing.B) {
	logng.Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.WithTime(testTime)
	}
}

func BenchmarkWithPrefix(b *testing.B) {
	logng.Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.WithPrefix("prefix")
	}
}

func BenchmarkWithPrefixf(b *testing.B) {
	logng.Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.WithPrefixf("%s", "prefix")
	}
}

func BenchmarkWithFieldKeyVals(b *testing.B) {
	logng.Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logng.WithFieldKeyVals("key1", "value1")
	}
}

func BenchmarkLogger_Info_withStackTrace(b *testing.B) {
	logger := logng.NewLogger(nopOutput{}, logng.SeverityInfo, 0)
	logger.SetStackTraceSeverity(logng.SeverityInfo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark")
	}
}

type nopOutput struct{}

func (nopOutput) Log(log *logng.Log) {}

var (
	testTime, _ = time.ParseInLocation("2006-01-02T15:04:05", "2010-11-12T13:14:15", time.Local)
)
