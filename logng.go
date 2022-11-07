package logng

import (
	"runtime"
)

// ProgramCounters returns program counters by using runtime.Callers.
func ProgramCounters(size, skip int) []uintptr {
	pc := make([]uintptr, size)
	pc = pc[:runtime.Callers(skip, pc)]
	return pc
}
