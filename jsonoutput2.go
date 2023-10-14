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

// JSONOutput2 is an implementation of Output by writing json to io.Writer w.
type JSONOutput2 struct {
	mu      sync.Mutex
	w       io.Writer
	flags   JSONOutput2Flag
	onError *func(error)
}

// NewJSONOutput2 creates a new JSONOutput2.
func NewJSONOutput2(w io.Writer, flags JSONOutput2Flag) *JSONOutput2 {
	return &JSONOutput2{
		w:     w,
		flags: flags,
	}
}

// Log is implementation of Output.
func (o *JSONOutput2) Log(log *Log) {
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
		Severity      *string    `json:"severity,omitempty"`
		Message       string     `json:"message"`
		Time          *time.Time `json:"time,omitempty"`
		Timestamp     *int64     `json:"timestamp,omitempty"`
		SeverityLevel *int       `json:"severityLevel,omitempty"`
		Verbosity     *int       `json:"verbosity,omitempty"`
		Func          *string    `json:"func,omitempty"`
		File          *string    `json:"file,omitempty"`
		StackTrace    *string    `json:"stack_trace,omitempty"`
	}
	data.Message = string(log.Message)

	if o.flags&JSONOutput2FlagSeverity != 0 {
		x := log.Severity.String()
		data.Severity = &x
	}

	if o.flags&(JSONOutput2FlagTime|JSONOutput2FlagTimestamp) != 0 {
		tm := log.Time
		if o.flags&JSONOutput2FlagUTC != 0 {
			tm = tm.UTC()
		}
		if o.flags&JSONOutput2FlagTime != 0 {
			data.Time = &tm
		}
		if o.flags&JSONOutput2FlagTimestamp != 0 {
			x := tm.Unix()
			data.Timestamp = &x
		}
	}

	if o.flags&JSONOutput2FlagSeverityLevel != 0 {
		x := int(log.Severity)
		data.SeverityLevel = &x
	}

	if o.flags&JSONOutput2FlagVerbosity != 0 {
		x := int(log.Verbosity)
		data.Verbosity = &x
	}

	if o.flags&(JSONOutput2FlagLongFunc|JSONOutput2FlagShortFunc) != 0 {
		fn := "???"
		if log.StackCaller.Function != "" {
			fn = trimSrcPath(log.StackCaller.Function)
		}
		if o.flags&JSONOutput2FlagShortFunc != 0 {
			fn = trimDirs(fn)
		}
		data.Func = &fn
	}

	if o.flags&(JSONOutput2FlagLongFile|JSONOutput2FlagShortFile) != 0 {
		file, line := "???", 0
		if log.StackCaller.File != "" {
			file = trimSrcPath(log.StackCaller.File)
			if o.flags&JSONOutput2FlagShortFile != 0 {
				file = trimDirs(file)
			}
		}
		if log.StackCaller.Line > 0 {
			line = log.StackCaller.Line
		}
		x := fmt.Sprintf("%s:%d", file, line)
		data.File = &x
	}

	if o.flags&JSONOutput2FlagStackTrace != 0 && log.StackTrace != nil {
		x := fmt.Sprintf("%+.1s", log.StackTrace)
		data.StackTrace = &x
	}

	fieldsKvs := make([]string, 0, 2*len(log.Fields))
	if o.flags&JSONOutput2FlagFields != 0 {
		fieldsMap := make(map[string]string, len(log.Fields))
		for idx, field := range log.Fields {
			key := fmt.Sprintf("_%s", field.Key)
			if _, ok := fieldsMap[key]; ok {
				key = fmt.Sprintf("%d_%s", idx, field.Key)
			}
			val := fmt.Sprintf("%v", field.Value)
			fieldsMap[key] = val
			fieldsKvs = append(fieldsKvs, key, val)
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

	for i, j := 0, len(fieldsKvs); i < j; i = i + 2 {
		buf.WriteRune(',')
		b, err = json.Marshal(map[string]string{fieldsKvs[i]: fieldsKvs[i+1]})
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
// It returns underlying JSONOutput2.
func (o *JSONOutput2) SetWriter(w io.Writer) *JSONOutput2 {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.w = w
	return o
}

// SetFlags sets flags to override every single Log.Flags if the argument flags different from 0.
// It returns underlying JSONOutput2.
// By default, 0.
func (o *JSONOutput2) SetFlags(flags JSONOutput2Flag) *JSONOutput2 {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.flags = flags
	return o
}

// SetOnError sets a function to call when error occurs.
// It returns underlying JSONOutput2.
func (o *JSONOutput2) SetOnError(f func(error)) *JSONOutput2 {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&o.onError)), unsafe.Pointer(&f))
	return o
}

type JSONOutput2Flag int

const (
	JSONOutput2FlagSeverity JSONOutput2Flag = 1 << iota

	JSONOutput2FlagTime

	JSONOutput2FlagTimestamp

	JSONOutput2FlagUTC

	JSONOutput2FlagSeverityLevel

	JSONOutput2FlagVerbosity

	JSONOutput2FlagLongFunc

	JSONOutput2FlagShortFunc

	JSONOutput2FlagLongFile

	JSONOutput2FlagShortFile

	JSONOutput2FlagStackTrace

	JSONOutput2FlagFields

	JSONOutput2FlagDefault = JSONOutput2FlagSeverity | JSONOutput2FlagTime | JSONOutput2FlagStackTrace | JSONOutput2FlagFields
)
