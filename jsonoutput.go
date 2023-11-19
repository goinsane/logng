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
	mu         sync.Mutex
	w          io.Writer
	flags      JSONOutputFlag
	onError    *func(error)
	timeLayout string
}

// NewJSONOutput creates a new JSONOutput.
func NewJSONOutput(w io.Writer, flags JSONOutputFlag) *JSONOutput {
	return &JSONOutput{
		w:          w,
		flags:      flags,
		timeLayout: time.RFC3339Nano,
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
		Severity      *string `json:"severity,omitempty"`
		Message       string  `json:"message"`
		Time          *string `json:"time,omitempty"`
		Timestamp     *int64  `json:"timestamp,omitempty"`
		SeverityLevel *int    `json:"severity_level,omitempty"`
		Verbosity     *int    `json:"verbosity,omitempty"`
		Func          *string `json:"func,omitempty"`
		File          *string `json:"file,omitempty"`
		StackTrace    *string `json:"stack_trace,omitempty"`
	}
	data.Message = string(log.Message)

	if o.flags&JSONOutputFlagSeverity != 0 {
		x := log.Severity.String()
		data.Severity = &x
	}

	if o.flags&(JSONOutputFlagTime|JSONOutputFlagTimestamp|JSONOutputFlagTimestampMicro) != 0 {
		tm := log.Time
		if o.flags&JSONOutputFlagUTC != 0 {
			tm = tm.UTC()
		}
		if o.flags&JSONOutputFlagTime != 0 {
			x := tm.Format(o.timeLayout)
			data.Time = &x
		}
		if o.flags&(JSONOutputFlagTimestamp|JSONOutputFlagTimestampMicro) != 0 {
			var x int64
			if o.flags&JSONOutputFlagTimestampMicro == 0 {
				x = tm.Unix()
			} else {
				x = tm.Unix()*1e6 + int64(tm.Nanosecond())/1e3
			}
			data.Timestamp = &x
		}
	}

	if o.flags&JSONOutputFlagSeverityLevel != 0 {
		x := int(log.Severity)
		data.SeverityLevel = &x
	}

	if o.flags&JSONOutputFlagVerbosity != 0 {
		x := int(log.Verbosity)
		data.Verbosity = &x
	}

	if o.flags&(JSONOutputFlagLongFunc|JSONOutputFlagShortFunc) != 0 {
		fn := "???"
		if log.StackCaller.Function != "" {
			fn = log.StackCaller.Function
		}
		if o.flags&JSONOutputFlagShortFunc != 0 {
			fn = trimDirs(fn)
		}
		data.Func = &fn
	}

	if o.flags&(JSONOutputFlagLongFile|JSONOutputFlagShortFile) != 0 {
		file, line := "???", 0
		if log.StackCaller.File != "" {
			file = log.StackCaller.File
			if o.flags&JSONOutputFlagShortFile != 0 {
				file = trimDirs(file)
			}
		}
		if log.StackCaller.Line > 0 {
			line = log.StackCaller.Line
		}
		x := fmt.Sprintf("%s:%d", file, line)
		data.File = &x
	}

	if o.flags&JSONOutputFlagStackTrace != 0 && log.StackTrace != nil {
		x := fmt.Sprintf("%+.1s", log.StackTrace)
		data.StackTrace = &x
	}

	fieldsKvs := make([]string, 0, 2*len(log.Fields))
	if o.flags&JSONOutputFlagFields != 0 {
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
// It returns the underlying JSONOutput.
func (o *JSONOutput) SetWriter(w io.Writer) *JSONOutput {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.w = w
	return o
}

// SetFlags sets flags to override every single Log.Flags if the argument flags different from 0.
// It returns the underlying JSONOutput.
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

// SetTimeLayout sets a time layout to format time field.
// It returns the underlying JSONOutput.
func (o *JSONOutput) SetTimeLayout(timeLayout string) *JSONOutput {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.timeLayout = timeLayout
	return o
}

type JSONOutputFlag int

const (
	JSONOutputFlagSeverity JSONOutputFlag = 1 << iota

	JSONOutputFlagTime

	JSONOutputFlagTimestamp

	JSONOutputFlagTimestampMicro

	JSONOutputFlagUTC

	JSONOutputFlagSeverityLevel

	JSONOutputFlagVerbosity

	JSONOutputFlagLongFunc

	JSONOutputFlagShortFunc

	JSONOutputFlagLongFile

	JSONOutputFlagShortFile

	JSONOutputFlagStackTrace

	JSONOutputFlagFields

	JSONOutputFlagDefault = JSONOutputFlagSeverity | JSONOutputFlagTime | JSONOutputFlagLongFunc |
		JSONOutputFlagShortFile | JSONOutputFlagStackTrace | JSONOutputFlagFields
)
