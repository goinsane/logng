package logng

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

// wrappedError is an interface to simulate GoLang's wrapped errors.
type wrappedError interface {
	error
	Unwrap() error
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

var (
	goRootSrcPath = filepath.Join(build.Default.GOROOT, "src") + string(os.PathSeparator)
	goSrcPath     = filepath.Join(build.Default.GOPATH, "src") + string(os.PathSeparator)
	sPkgModPath   = filepath.Join(build.Default.GOPATH, filepath.Join("pkg", "mod")) + string(os.PathSeparator)
)

func trimSrcPath(s string) string {
	var r string
	r = strings.TrimPrefix(s, goRootSrcPath)
	if r != s {
		return r
	}
	r = strings.TrimPrefix(s, goSrcPath)
	if r != s {
		return r
	}
	r = strings.TrimPrefix(s, sPkgModPath)
	if r != s {
		return r
	}
	return s
}

func trimDirs(s string) string {
	for i := len(s) - 1; i > 0; i-- {
		if s[i] == '/' || s[i] == os.PathSeparator {
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
