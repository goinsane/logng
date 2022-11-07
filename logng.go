package logng

import (
	"runtime"
)

// ProgramCounters returns program counters by using runtime.Callers.
func ProgramCounters(size, skip int) []uintptr {
	programCounter := make([]uintptr, size)
	programCounter = programCounter[:runtime.Callers(skip, programCounter)]
	return programCounter
}
