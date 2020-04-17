package op

import (
	"errors"
	"os"
	"syscall"
)

// ExitHinter is an interface that if implemented by an error type, will be
// used within the ExitHint function.
type ExitHinter interface {
	ExitHint() int
}

// ExitHint returns an exit code hint according to the passed in signal and
// error. On Unix systems, 128 + int(signal) is returned when err is not nil.
func ExitHint(signal os.Signal, err error) int {
	s, unixSignal := signal.(syscall.Signal)
	var eh ExitHinter
	switch {
	case err == nil:
		return 0
	case unixSignal:
		return 128 + int(s)
	case errors.As(err, &eh):
		return eh.ExitHint()
	default:
		return 1
	}
}
