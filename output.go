package logng

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Output is an interface for Logger output.
// All the Output implementations must be safe for concurrency.
type Output interface {
	Log(log *Log)
}

type multiOutput []Output

func (m multiOutput) Log(log *Log) {
	for _, o := range m {
		o.Log(log.Clone())
	}
}

// MultiOutput creates an output that clones its logs to all the provided outputs.
func MultiOutput(outputs ...Output) Output {
	m := make(multiOutput, len(outputs))
	copy(m, outputs)
	return m
}

// QueuedOutput is intermediate Output implementation between Logger and given Output.
// QueuedOutput has queueing for unblocking Log() method.
type QueuedOutput struct {
	output      Output
	queue       chan *Log
	ctx         context.Context
	ctxCancel   context.CancelFunc
	blocking    uint32
	onQueueFull *func()
}

// NewQueuedOutput creates QueuedOutput by given output.
func NewQueuedOutput(output Output, queueLen int) (q *QueuedOutput) {
	if queueLen < 0 {
		queueLen = 0
	}
	q = &QueuedOutput{
		output: output,
		queue:  make(chan *Log, queueLen),
	}
	q.ctx, q.ctxCancel = context.WithCancel(context.Background())
	go q.worker()
	return
}

// Close closed QueuedOutput. Unused QueuedOutput must be closed for freeing resources.
func (q *QueuedOutput) Close() error {
	q.ctxCancel()
	return nil
}

// Log is implementation of Output.
// If blocking is true, Log method blocks execution until underlying output has finished execution.
// Otherwise, Log method sends log to queue if queue is available. When queue is full, it tries to call OnQueueFull
// function.
func (q *QueuedOutput) Log(log *Log) {
	select {
	case <-q.ctx.Done():
		return
	default:
	}
	if q.blocking != 0 {
		q.queue <- log
		return
	}
	select {
	case q.queue <- log:
	default:
		if q.onQueueFull != nil && *q.onQueueFull != nil {
			(*q.onQueueFull)()
		}
	}
}

// SetBlocking sets QueuedOutput behavior when queue is full.
// It returns underlying QueuedOutput.
func (q *QueuedOutput) SetBlocking(blocking bool) *QueuedOutput {
	var b uint32
	if blocking {
		b = 1
	}
	atomic.StoreUint32(&q.blocking, b)
	return q
}

// SetOnQueueFull sets a function to call when queue is full.
// It returns underlying QueuedOutput.
func (q *QueuedOutput) SetOnQueueFull(f func()) *QueuedOutput {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&q.onQueueFull)), unsafe.Pointer(&f))
	return q
}

// WaitForEmpty waits until queue is empty by given context.
func (q *QueuedOutput) WaitForEmpty(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			if len(q.queue) == 0 {
				return nil
			}
		}
	}
}

func (q *QueuedOutput) worker() {
	for done := false; !done; {
		select {
		case <-q.ctx.Done():
			done = true
		case msg := <-q.queue:
			if q.output != nil {
				q.output.Log(msg)
			}
		}
	}
}

// TextOutput is an implementation of Output by writing texts to io.Writer w.
type TextOutput struct {
	mu      sync.Mutex
	w       io.Writer
	flags   TextOutputFlag
	onError *func(error)
}

// NewTextOutput creates a new TextOutput.
func NewTextOutput(w io.Writer, flags TextOutputFlag) *TextOutput {
	return &TextOutput{
		w:     w,
		flags: flags,
	}
}

// Log is implementation of Output.
func (t *TextOutput) Log(log *Log) {
	var err error
	defer func() {
		if err == nil || t.onError == nil || *t.onError == nil {
			return
		}
		(*t.onError)(err)
	}()

	t.mu.Lock()
	defer t.mu.Unlock()

	buf := bytes.NewBuffer(make([]byte, 0, 4096))

	if t.flags&(TextOutputFlagDate|TextOutputFlagTime|TextOutputFlagMicroseconds) != 0 {
		tm := log.Time.Local()
		if t.flags&TextOutputFlagUTC != 0 {
			tm = tm.UTC()
		}
		b := make([]byte, 0, 128)
		if t.flags&TextOutputFlagDate != 0 {
			year, month, day := tm.Date()
			itoa(&b, year, 4)
			b = append(b, '/')
			itoa(&b, int(month), 2)
			b = append(b, '/')
			itoa(&b, day, 2)
			b = append(b, ' ')
		}
		if t.flags&(TextOutputFlagTime|TextOutputFlagMicroseconds) != 0 {
			hour, min, sec := tm.Clock()
			itoa(&b, hour, 2)
			b = append(b, ':')
			itoa(&b, min, 2)
			b = append(b, ':')
			itoa(&b, sec, 2)
			if t.flags&TextOutputFlagMicroseconds != 0 {
				b = append(b, '.')
				itoa(&b, log.Time.Nanosecond()/1e3, 6)
			}
			b = append(b, ' ')
		}
		buf.Write(b)
	}

	if t.flags&TextOutputFlagSeverity != 0 {
		buf.WriteString(log.Severity.String())
		buf.WriteString(" - ")
	}

	var padding []byte
	if t.flags&TextOutputFlagPadding != 0 {
		padding = bytes.Repeat([]byte(" "), buf.Len())
	}

	if t.flags&(TextOutputFlagLongFunc|TextOutputFlagShortFunc) != 0 {
		fn := "???"
		if log.StackCaller.Function != "" {
			fn = trimSrcPath(log.StackCaller.Function)
		}
		if t.flags&TextOutputFlagShortFunc != 0 {
			fn = trimDirs(fn)
		}
		buf.WriteString(fn)
		buf.WriteString("()")
		buf.WriteString(" - ")
	}

	if t.flags&(TextOutputFlagLongFile|TextOutputFlagShortFile) != 0 {
		file, line := "???", 0
		if log.StackCaller.File != "" {
			file = trimSrcPath(log.StackCaller.File)
			if t.flags&TextOutputFlagShortFile != 0 {
				file = trimDirs(file)
			}
		}
		if log.StackCaller.Line > 0 {
			line = log.StackCaller.Line
		}
		buf.WriteString(file)
		buf.WriteRune(':')
		b := make([]byte, 0, 128)
		itoa(&b, line, -1)
		buf.Write(b)
		buf.WriteString(" - ")
	}

	for idx, line := range bytes.Split(log.Message, []byte("\n")) {
		if idx > 0 {
			buf.Write(padding)
		}
		buf.Write(line)
		buf.WriteRune('\n')
	}

	extended := false
	extend := func() {
		if !extended {
			extended = true
			buf.WriteString("\t\n")
		}
	}

	if t.flags&TextOutputFlagFields != 0 && len(log.Fields) > 0 {
		extend()
		buf.WriteRune('\t')
		buf.WriteString("+ ")
		for idx, field := range log.Fields {
			if idx > 0 {
				buf.WriteRune(' ')
			}
			buf.WriteString(fmt.Sprintf("%q=%q", field.Key, fmt.Sprintf("%v", field.Value)))
		}
		buf.WriteString("\n\t")
		buf.WriteRune('\n')
	}

	if t.flags&TextOutputFlagStackTrace != 0 && log.StackTrace != nil {
		extend()
		buf.WriteString(fmt.Sprintf("%+1.1s", log.StackTrace))
		buf.WriteString("\n\t")
		buf.WriteRune('\n')
	}

	_, err = t.w.Write(buf.Bytes())
	if err != nil {
		return
	}
}

// SetWriter sets writer.
// It returns underlying TextOutput.
func (t *TextOutput) SetWriter(w io.Writer) *TextOutput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.w = w
	return t
}

// SetFlags sets flags to override every single Log.Flags if the argument flags different from 0.
// It returns underlying TextOutput.
// By default, 0.
func (t *TextOutput) SetFlags(flags TextOutputFlag) *TextOutput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.flags = flags
	return t
}

// SetOnError sets a function to call when error occurs.
// It returns underlying TextOutput.
func (t *TextOutput) SetOnError(f func(error)) *TextOutput {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&t.onError)), unsafe.Pointer(&f))
	return t
}

// TextOutputFlag holds single or multiple flags of TextOutput.
// An TextOutput instance uses these flags which are stored by TextOutputFlag type.
type TextOutputFlag int

const (
	// TextOutputFlagDate prints the date in the local time zone: 2009/01/23
	TextOutputFlagDate TextOutputFlag = 1 << iota

	// TextOutputFlagTime prints the time in the local time zone: 01:23:23
	TextOutputFlagTime

	// TextOutputFlagMicroseconds prints microsecond resolution: 01:23:23.123123
	TextOutputFlagMicroseconds

	// TextOutputFlagUTC uses UTC rather than the local time zone
	TextOutputFlagUTC

	// TextOutputFlagSeverity prints severity level
	TextOutputFlagSeverity

	// TextOutputFlagPadding prints padding with multiple lines
	TextOutputFlagPadding

	// TextOutputFlagLongFunc prints full package name and function name: a/b/c/d.Func1()
	TextOutputFlagLongFunc

	// TextOutputFlagShortFunc prints final package name and function name: d.Func1()
	TextOutputFlagShortFunc

	// TextOutputFlagLongFile prints full file name and line number: a/b/c/d.go:23
	TextOutputFlagLongFile

	// TextOutputFlagShortFile prints final file name element and line number: d.go:23
	TextOutputFlagShortFile

	// TextOutputFlagFields prints fields if there are
	TextOutputFlagFields

	// TextOutputFlagStackTrace prints the stack trace if there is
	TextOutputFlagStackTrace

	// TextOutputFlagDefault holds initial flags for the Logger
	TextOutputFlagDefault = TextOutputFlagDate | TextOutputFlagTime | TextOutputFlagSeverity | TextOutputFlagPadding | TextOutputFlagFields | TextOutputFlagStackTrace
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

// Log is implementation of Output.
func (o *JSONOutput) Log(log *Log) {
	var err error
	defer func() {
		if err == nil || o.onError == nil || *o.onError == nil {
			return
		}
		(*o.onError)(err)
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
// It returns underlying JSONOutput.
func (o *JSONOutput) SetWriter(w io.Writer) *JSONOutput {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.w = w
	return o
}

// SetFlags sets flags to override every single Log.Flags if the argument flags different from 0.
// It returns underlying JSONOutput.
// By default, 0.
func (o *JSONOutput) SetFlags(flags JSONOutputFlag) *JSONOutput {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.flags = flags
	return o
}

// SetOnError sets a function to call when error occurs.
// It returns underlying JSONOutput.
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
