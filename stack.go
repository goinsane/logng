package logng

import (
	"bytes"
	"fmt"
	"runtime"
)

// StackCaller stores the information of the stack caller.
// StackCaller can format the given information as string by using String or Format methods.
type StackCaller struct {
	runtime.Frame
}

// String is the implementation of fmt.Stringer.
// It is synonym with fmt.Sprintf("%s", c).
func (c StackCaller) String() string {
	return fmt.Sprintf("%s", c)
}

// Format is the implementation of fmt.Formatter.
//
// For '%s' (also '%v'):
//
//	%s       just show function and entry without padding and indent.
//	%+s      show file path, line and program counter. padding char '\t', default padding 0, default indent 1.
//	% s      show file path, line and program counter. padding char ' ', default padding 0, default indent 2.
//	%+ s     exact with '% s'.
//	%#s      same with '%+s', use file name as file path.
//	%+#s     exact with '%#s'.
//	% #s     same with '% s', use file name as file path.
//	%+4s     same with '%+s', padding 4, indent 1 by default.
//	%+.3s    same with '%+s', padding 0 by default, indent 3.
//	%+4.3s   same with '%+s', padding 4, indent 3.
//	%+4.s    same with '%+s', padding 4, indent 0.
//	% 4s     same with '% s', padding 4, indent 2 by default.
//	% .3s    same with '% s', padding 0 by default, indent 3.
//	% 4.3s   same with '% s', padding 4, indent 3.
//	% 4.s    same with '% s', padding 4, indent 0.
//	%#4.3s   same with '%#s', padding 4, indent 3.
//	% #4.3s  same with '% #s', padding 4, indent 3.
func (c StackCaller) Format(f fmt.State, verb rune) {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	switch verb {
	case 's', 'v':
		fn := "???"
		if c.Function != "" {
			fn = c.Function
		}
		extended := f.Flag('+') || f.Flag(' ') || f.Flag('#')
		if !extended {
			buf.WriteString(fmt.Sprintf("%s(%#x)", fn, c.Entry))
			break
		}
		pad, wid, prec := getPadWidPrec(f)
		padding, indent := bytes.Repeat([]byte{pad}, wid), bytes.Repeat([]byte{pad}, prec)
		buf.Write(padding)
		buf.WriteString(fmt.Sprintf("%s(%#x)", fn, c.Entry))
		buf.WriteRune('\n')
		buf.Write(padding)
		buf.Write(indent)
		file, line := "???", 0
		if c.File != "" {
			file = c.File
			if f.Flag('#') {
				file = trimDirs(file)
			}
		}
		if c.Line > 0 {
			line = c.Line
		}
		buf.WriteString(fmt.Sprintf("%s:%d +%#x", file, line, c.PC-c.Entry))
	default:
		return
	}
	_, _ = f.Write(buf.Bytes())
}

// StackTrace stores the information of the stack trace.
type StackTrace struct {
	programCounters []uintptr
	callers         []StackCaller
}

// NewStackTrace creates a new StackTrace.
func NewStackTrace(programCounters []uintptr) *StackTrace {
	t := &StackTrace{
		programCounters: make([]uintptr, len(programCounters)),
		callers:         make([]StackCaller, 0, len(programCounters)),
	}
	copy(t.programCounters, programCounters)
	if len(t.programCounters) > 0 {
		frames := runtime.CallersFrames(t.programCounters)
		for {
			frame, more := frames.Next()
			caller := StackCaller{
				Frame: frame,
			}
			t.callers = append(t.callers, caller)
			if !more {
				break
			}
		}
	}
	return t
}

// Clone clones the underlying StackTrace.
func (t *StackTrace) Clone() *StackTrace {
	if t == nil {
		return nil
	}
	t2 := &StackTrace{
		programCounters: make([]uintptr, len(t.programCounters), cap(t.programCounters)),
		callers:         make([]StackCaller, len(t.callers), cap(t.callers)),
	}
	copy(t2.programCounters, t.programCounters)
	copy(t2.callers, t.callers)
	return t2
}

// String is the implementation of fmt.Stringer.
// It is synonym with fmt.Sprintf("%s", t).
func (t *StackTrace) String() string {
	return fmt.Sprintf("%s", t)
}

// Format is the implementation of fmt.Formatter.
// Format lists all StackCaller's in the underlying StackTrace line by line with the given format.
func (t *StackTrace) Format(f fmt.State, verb rune) {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	switch verb {
	case 's', 'v':
		format := "%"
		for _, r := range []rune{'+', ' ', '#'} {
			if f.Flag(int(r)) {
				format += string(r)
			}
		}
		_, wid, prec := getPadWidPrec(f)
		format += fmt.Sprintf("%d.%ds", wid, prec)
		for i, c := range t.callers {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(fmt.Sprintf(format, c))
		}
	default:
		return
	}
	_, _ = f.Write(buf.Bytes())
}

// ProgramCounters returns program counters.
func (t *StackTrace) ProgramCounters() []uintptr {
	result := make([]uintptr, len(t.programCounters))
	copy(result, t.programCounters)
	return result
}

// SizeOfProgramCounters returns the size of program counters.
func (t *StackTrace) SizeOfProgramCounters() int {
	return len(t.programCounters)
}

// Callers returns StackCaller's.
func (t *StackTrace) Callers() []StackCaller {
	result := make([]StackCaller, len(t.callers))
	copy(result, t.callers)
	return result
}

// SizeOfCallers returns the size of StackCaller's.
func (t *StackTrace) SizeOfCallers() int {
	return len(t.callers)
}

// Caller returns a caller on the given index. It panics if index is out of range.
func (t *StackTrace) Caller(index int) StackCaller {
	if 0 > index || index >= len(t.callers) {
		panic("index out of range")
	}
	return t.callers[index]
}

// ProgramCounters returns program counters by using runtime.Callers.
func ProgramCounters(size, skip int) []uintptr {
	programCounter := make([]uintptr, size)
	programCounter = programCounter[:runtime.Callers(skip, programCounter)]
	return programCounter
}
