package main

import (
	"fmt"

	"github.com/goinsane/logng"
)

func main() {
	s := logng.NewStackTrace(logng.ProgramCounters(1, 0))
	fmt.Printf("%+s\n", s)
}
