package logng

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Output is an interface for Logger output.
// All of Output implementations must be safe for concurrency.
type Output interface {
	Log(log *Log)
}

type multiOutput []Output

func (o multiOutput) Log(log *Log) {
	for _, o1 := range o {
		o1.Log(log)
	}
}

// MultiOutput creates an output that clones its logs to all the provided outputs.
func MultiOutput(outputs ...Output) Output {
	o := make(multiOutput, len(outputs))
	copy(o, outputs)
	return o
}

// QueuedOutput is intermediate Output implementation between Logger and given Output.
// QueuedOutput has queueing for unblocking Log() method.
type QueuedOutput struct {
	output      Output
	queue       chan *Log
	closing     int32
	wg          sync.WaitGroup
	logWg       sync.WaitGroup
	blocking    uint32
	onQueueFull *func()
}

// NewQueuedOutput creates a new QueuedOutput by the given output.
func NewQueuedOutput(output Output, queueLen int) (o *QueuedOutput) {
	o = &QueuedOutput{
		output: output,
		queue:  make(chan *Log, queueLen),
	}
	o.wg.Add(1)
	go o.worker()
	return
}

// Close stops accepting new logs to the underlying QueuedOutput and waits for the queue to empty.
// Unused QueuedOutput must be closed for freeing resources.
func (o *QueuedOutput) Close() error {
	if !atomic.CompareAndSwapInt32(&o.closing, 0, 1) {
		return nil
	}
	o.logWg.Wait()
	close(o.queue)
	o.wg.Wait()
	return nil
}

// Log is the implementation of Output.
// If blocking is true, Log method blocks execution until the underlying output has finished execution.
// Otherwise, Log method sends the log to the queue if the queue is available.
// When the queue is full, it tries to call OnQueueFull function.
func (o *QueuedOutput) Log(log *Log) {
	o.logWg.Add(1)
	defer o.logWg.Done()
	if o.closing != 0 {
		return
	}
	if o.blocking != 0 {
		o.queue <- log
		return
	}
	select {
	case o.queue <- log:
	default:
		onQueueFull := o.onQueueFull
		if onQueueFull != nil && *onQueueFull != nil {
			(*onQueueFull)()
		}
	}
}

// SetBlocking sets QueuedOutput behavior when the queue is full.
// It returns the underlying QueuedOutput.
func (o *QueuedOutput) SetBlocking(blocking bool) *QueuedOutput {
	var b uint32
	if blocking {
		b = 1
	}
	atomic.StoreUint32(&o.blocking, b)
	return o
}

// SetOnQueueFull sets a function to call when the queue is full.
// It returns the underlying QueuedOutput.
func (o *QueuedOutput) SetOnQueueFull(f func()) *QueuedOutput {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&o.onQueueFull)), unsafe.Pointer(&f))
	return o
}

func (o *QueuedOutput) worker() {
	defer o.wg.Done()
	for log := range o.queue {
		o.output.Log(log)
	}
}

// TextOutput is an implementation of Output by writing texts to io.Writer w.
type TextOutput struct {
	mu      sync.RWMutex
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

// Log is the implementation of Output.
func (o *TextOutput) Log(log *Log) {
	var err error
	defer func() {
		onError := o.onError
		if err == nil || onError == nil || *onError == nil {
			return
		}
		(*onError)(err)
	}()

	o.mu.RLock()
	defer o.mu.RUnlock()

	buf := bytes.NewBuffer(make([]byte, 0, 4096))

	if o.flags&(TextOutputFlagDate|TextOutputFlagTime|TextOutputFlagMicroseconds) != 0 {
		tm := log.Time.Local()
		if o.flags&TextOutputFlagUTC != 0 {
			tm = tm.UTC()
		}
		b := make([]byte, 0, 128)
		if o.flags&TextOutputFlagDate != 0 {
			year, month, day := tm.Date()
			itoa(&b, year, 4)
			b = append(b, '/')
			itoa(&b, int(month), 2)
			b = append(b, '/')
			itoa(&b, day, 2)
			b = append(b, ' ')
		}
		if o.flags&(TextOutputFlagTime|TextOutputFlagMicroseconds) != 0 {
			hour, min, sec := tm.Clock()
			itoa(&b, hour, 2)
			b = append(b, ':')
			itoa(&b, min, 2)
			b = append(b, ':')
			itoa(&b, sec, 2)
			if o.flags&TextOutputFlagMicroseconds != 0 {
				b = append(b, '.')
				itoa(&b, log.Time.Nanosecond()/1e3, 6)
			}
			b = append(b, ' ')
		}
		buf.Write(b)
	}

	if o.flags&TextOutputFlagSeverity != 0 {
		buf.WriteString(log.Severity.String())
		buf.WriteString(" - ")
	}

	var padding []byte
	if o.flags&TextOutputFlagPadding != 0 {
		padding = bytes.Repeat([]byte(" "), buf.Len())
	}

	if o.flags&(TextOutputFlagLongFunc|TextOutputFlagShortFunc) != 0 {
		fn := "???"
		if log.StackCaller.Function != "" {
			fn = log.StackCaller.Function
		}
		if o.flags&TextOutputFlagShortFunc != 0 {
			fn = trimDirs(fn)
		}
		buf.WriteString(fn)
		buf.WriteString("()")
		buf.WriteString(" - ")
	}

	if o.flags&(TextOutputFlagLongFile|TextOutputFlagShortFile) != 0 {
		file, line := "???", 0
		if log.StackCaller.File != "" {
			file = log.StackCaller.File
			if o.flags&TextOutputFlagShortFile != 0 {
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

	if o.flags&TextOutputFlagFields != 0 && len(log.Fields) > 0 {
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

	if o.flags&(TextOutputFlagStackTrace|TextOutputFlagStackTraceShortFile) != 0 && log.StackTrace != nil {
		extend()
		f := "%+1.1s"
		if o.flags&TextOutputFlagStackTraceShortFile != 0 {
			f = "%+#1.1s"
		}
		buf.WriteString(fmt.Sprintf(f, log.StackTrace))
		buf.WriteString("\n\t")
		buf.WriteRune('\n')
	}

	_, err = io.Copy(o.w, buf)
	if err != nil {
		err = fmt.Errorf("unable to write to writer: %w", err)
		return
	}
}

// SetWriter sets writer.
// It returns the underlying TextOutput.
func (o *TextOutput) SetWriter(w io.Writer) *TextOutput {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.w = w
	return o
}

// SetFlags sets flags to override every single Log.Flags if argument flags is different from 0.
// It returns the underlying TextOutput.
func (o *TextOutput) SetFlags(flags TextOutputFlag) *TextOutput {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.flags = flags
	return o
}

// SetOnError sets a function to call when error occurs.
// It returns the underlying TextOutput.
func (o *TextOutput) SetOnError(f func(error)) *TextOutput {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&o.onError)), unsafe.Pointer(&f))
	return o
}

// TextOutputFlag holds single or multiple flags of TextOutput.
// A TextOutput instance uses these flags which are stored by TextOutputFlag type.
type TextOutputFlag int

const (
	// TextOutputFlagDate prints the date in the local time zone: 2009/01/23.
	TextOutputFlagDate TextOutputFlag = 1 << iota

	// TextOutputFlagTime prints the time in the local time zone: 01:23:23.
	TextOutputFlagTime

	// TextOutputFlagMicroseconds prints microsecond resolution: 01:23:23.123123.
	// assumes TextOutputFlagTime.
	TextOutputFlagMicroseconds

	// TextOutputFlagUTC uses UTC rather than the local time zone if TextOutputFlagDate or TextOutputFlagTime is set.
	TextOutputFlagUTC

	// TextOutputFlagSeverity prints the severity.
	TextOutputFlagSeverity

	// TextOutputFlagPadding prints padding with multiple lines.
	TextOutputFlagPadding

	// TextOutputFlagLongFunc prints full package name and function name: a/b/c/d.Func1().
	TextOutputFlagLongFunc

	// TextOutputFlagShortFunc prints final package name and function name: d.Func1().
	// overrides TextOutputFlagLongFunc.
	TextOutputFlagShortFunc

	// TextOutputFlagLongFile prints full file name and line number: a/b/c/d.go:23.
	TextOutputFlagLongFile

	// TextOutputFlagShortFile prints final file name element and line number: d.go:23.
	// overrides TextOutputFlagLongFile.
	TextOutputFlagShortFile

	// TextOutputFlagFields prints fields if given.
	TextOutputFlagFields

	// TextOutputFlagStackTrace prints the stack trace if given.
	TextOutputFlagStackTrace

	// TextOutputFlagStackTraceShortFile prints with file name element only.
	// assumes TextOutputFlagStackTrace.
	TextOutputFlagStackTraceShortFile

	// TextOutputFlagDefault holds predefined default flags.
	// it used by the default Logger.
	TextOutputFlagDefault = TextOutputFlagDate | TextOutputFlagTime | TextOutputFlagSeverity |
		TextOutputFlagPadding | TextOutputFlagFields | TextOutputFlagStackTraceShortFile
)
