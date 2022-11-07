package logng

import (
	"fmt"
	"go/build"
	"os"
	"strings"
)

func trimSrcPath(s string) string {
	var r string
	r = strings.TrimPrefix(s, build.Default.GOROOT+"/src/")
	if r != s {
		return r
	}
	r = strings.TrimPrefix(s, build.Default.GOPATH+"/src/")
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
