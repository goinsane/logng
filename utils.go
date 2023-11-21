package logng

import (
	"fmt"
	"runtime"
)

// wrappedError is an interface to simulate GoLang's wrapped errors.
type wrappedError interface {
	error
	Unwrap() error
}

// programCounters returns program counters by using runtime.Callers.
func programCounters(size, skip int) []uintptr {
	pc := make([]uintptr, size)
	pc = pc[:runtime.Callers(skip, pc)]
	return pc
}

func itoa(buf *[]byte, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func trimDirs(s string) string {
	for i := len(s) - 1; i > 0; i-- {
		if s[i] == '/' {
			return s[i+1:]
		}
	}
	return s
}

func getPadWidPrec(f fmt.State) (pad byte, wid, prec int) {
	pad, wid, prec = byte('\t'), 0, 1
	if f.Flag(' ') {
		pad = ' '
		prec = 2
	}
	if w, ok := f.Width(); ok {
		wid = w
	}
	if p, ok := f.Precision(); ok {
		prec = p
	}
	return
}
