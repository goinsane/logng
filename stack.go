package logng

import (
	"bytes"
	"fmt"
	"runtime"
)

// StackCaller stores the information of stack caller.
// StackCaller can format given information as string by using Format or String methods.
type StackCaller struct {
	runtime.Frame
}

// String is implementation of fmt.Stringer.
// It is synonym with fmt.Sprintf("%s", c).
func (c StackCaller) String() string {
	return fmt.Sprintf("%s", c)
}

// Format is implementation of fmt.Formatter.
//
// For '%s' (also '%v'):
// 	%s       just show function and entry without padding and indent.
// 	%+s      show file path, line and pc. padding char '\t', default padding 0, default indent 1.
// 	% s      show file path, line and pc. padding char ' ', default padding 0, default indent 2.
// 	%+ s     exact with '% s'.
// 	%#s      same with '%+s', use file name as file path.
// 	%+#s     exact with '%#s'.
// 	% #s     same with '% s', use file name as file path.
// 	%+4s     same with '%+s', padding 4, indent 1 by default.
// 	%+.3s    same with '%+s', padding 0 by default, indent 3.
// 	%+4.3s   same with '%+s', padding 4, indent 3.
// 	%+4.s    same with '%+s', padding 4, indent 0.
// 	% 4s     same with '% s', padding 4, indent 2 by default.
// 	% .3s    same with '% s', padding 0 by default, indent 3.
// 	% 4.3s   same with '% s', padding 4, indent 3.
// 	% 4.s    same with '% s', padding 4, indent 0.
// 	%#4.3s   same with '%#s', padding 4, indent 3.
// 	% #4.3s  same with '% #s', padding 4, indent 3.
func (c StackCaller) Format(f fmt.State, verb rune) {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	switch verb {
	case 's', 'v':
		fn := "???"
		if c.Function != "" {
			fn = trimSrcPath(c.Function)
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
			file = trimSrcPath(c.File)
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

// StackTrace stores the information of stack trace.
type StackTrace struct {
	pc      []uintptr
	callers []StackCaller
}

// NewStackTrace creates a new StackTrace object.
func NewStackTrace(pc ...uintptr) *StackTrace {
	t := &StackTrace{
		pc:      make([]uintptr, len(pc)),
		callers: make([]StackCaller, 0, len(pc)),
	}
	copy(t.pc, pc)
	if len(t.pc) > 0 {
		frames := runtime.CallersFrames(t.pc)
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

// Clone clones the StackTrace object.
func (t *StackTrace) Clone() *StackTrace {
	if t == nil {
		return nil
	}
	t2 := &StackTrace{
		pc:      make([]uintptr, len(t.pc), cap(t.pc)),
		callers: make([]StackCaller, len(t.callers), cap(t.callers)),
	}
	copy(t2.pc, t.pc)
	copy(t2.callers, t.callers)
	return t2
}

// String is implementation of fmt.Stringer.
// It is synonym with fmt.Sprintf("%s", t).
func (t *StackTrace) String() string {
	return fmt.Sprintf("%s", t)
}

// Format is implementation of fmt.Formatter.
// Format lists all StackCaller's in StackTrace line by line with given format.
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

// PC returns program counters.
func (t *StackTrace) PC() []uintptr {
	result := make([]uintptr, len(t.pc))
	copy(result, t.pc)
	return result
}

// Caller returns a StackCaller on the given index. It panics if index is out of range.
func (t *StackTrace) Caller(index int) StackCaller {
	if index < 0 || index >= len(t.callers) {
		panic("index out of range")
	}
	return t.callers[index]
}

// Len returns the length of the length of all Caller's.
func (t *StackTrace) Len() int {
	return len(t.callers)
}
