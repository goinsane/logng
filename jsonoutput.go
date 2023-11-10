package logng

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// JSONOutput is an implementation of Output by writing json to io.Writer w.
type JSONOutput struct {
	mu      sync.Mutex
	w       io.Writer
	flags   JSONOutputFlag
	onError *func(error)
}

// NewJSONOutput creates a new JSONOutput.
func NewJSONOutput(w io.Writer, flags JSONOutputFlag) *JSONOutput {
	return &JSONOutput{
		w:     w,
		flags: flags,
	}
}

// Log is the implementation of Output.
func (o *JSONOutput) Log(log *Log) {
	var err error
	defer func() {
		onError := o.onError
		if err == nil || onError == nil || *onError == nil {
			return
		}
		(*onError)(err)
	}()

	o.mu.Lock()
	defer o.mu.Unlock()

	buf := bytes.NewBuffer(make([]byte, 0, 4096))

	var data struct {
		Time           time.Time         `json:"time"`
		SeverityString string            `json:"severity,omitempty"`
		Message        string            `json:"message"`
		Severity       int               `json:"s"`
		Verbosity      int               `json:"v"`
		Func           string            `json:"func,omitempty"`
		File           string            `json:"file,omitempty"`
		StackTrace     string            `json:"stack_trace,omitempty"`
		Fields         map[string]string `json:"-"`
	}

	data.Time = log.Time
	data.Message = string(log.Message)
	data.Severity = int(log.Severity)
	data.Verbosity = int(log.Verbosity)

	if o.flags&JSONOutputFlagUTC != 0 {
		data.Time = data.Time.UTC()
	}

	if o.flags&JSONOutputFlagSeverity != 0 {
		data.SeverityString = log.Severity.String()
	}

	if o.flags&JSONOutputFlagFunc != 0 {
		fn := "???"
		if log.StackCaller.Function != "" {
			fn = trimSrcPath(log.StackCaller.Function)
		}
		data.Func = fn
	}

	if o.flags&JSONOutputFlagFile != 0 {
		file, line := "???", 0
		if log.StackCaller.File != "" {
			file = trimSrcPath(log.StackCaller.File)
		}
		if log.StackCaller.Line > 0 {
			line = log.StackCaller.Line
		}
		data.File = fmt.Sprintf("%s:%d", file, line)
	}

	if o.flags&JSONOutputFlagStackTrace != 0 && log.StackTrace != nil {
		data.StackTrace = fmt.Sprintf("%+.1s", log.StackTrace)
	}

	if o.flags&JSONOutputFlagFields != 0 {
		data.Fields = make(map[string]string, 4096)
		for idx, field := range log.Fields {
			key := fmt.Sprintf("_%s", field.Key)
			if _, ok := data.Fields[key]; ok {
				key = fmt.Sprintf("%d_%s", idx, field.Key)
			}
			data.Fields[key] = fmt.Sprintf("%v", field.Value)
		}
	}

	buf.WriteRune('{')
	var b []byte

	b, err = json.Marshal(&data)
	if err != nil {
		return
	}
	b = bytes.TrimLeft(b, "{")
	b = bytes.TrimRight(b, "}")
	buf.Write(b)

	if len(data.Fields) > 0 {
		buf.WriteRune(',')
		b, err = json.Marshal(data.Fields)
		if err != nil {
			return
		}
		b = bytes.TrimLeft(b, "{")
		b = bytes.TrimRight(b, "}")
		buf.Write(b)
	}

	buf.WriteRune('}')
	buf.WriteRune('\n')

	_, err = o.w.Write(buf.Bytes())
	if err != nil {
		return
	}
}

// SetWriter sets writer.
// It returns the underlying JSONOutput.
func (o *JSONOutput) SetWriter(w io.Writer) *JSONOutput {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.w = w
	return o
}

// SetFlags sets flags to override every single Log.Flags if the argument flags different from 0.
// It returns the underlying JSONOutput.
// By default, 0.
func (o *JSONOutput) SetFlags(flags JSONOutputFlag) *JSONOutput {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.flags = flags
	return o
}

// SetOnError sets a function to call when error occurs.
// It returns the underlying JSONOutput.
func (o *JSONOutput) SetOnError(f func(error)) *JSONOutput {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&o.onError)), unsafe.Pointer(&f))
	return o
}

type JSONOutputFlag int

const (
	JSONOutputFlagUTC JSONOutputFlag = 1 << iota

	JSONOutputFlagSeverity

	JSONOutputFlagFunc

	JSONOutputFlagFile

	JSONOutputFlagStackTrace

	JSONOutputFlagFields

	JSONOutputFlagDefault = JSONOutputFlagSeverity | JSONOutputFlagFunc | JSONOutputFlagFile | JSONOutputFlagStackTrace | JSONOutputFlagFields
)
