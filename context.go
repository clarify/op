package op

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ctxKey uint8

const (
	opKey ctxKey = 0
)

// CancelFunc is a function that cancel a running context before returning an
// os.Signal if one has been received. After being called once, the function
// will continue to return the same response. Multiple concurrent calls are
// safe.
type CancelFunc = func() os.Signal

// closedchan is a reusable closed channel.
var closedchan = make(chan struct{})

func init() {
	close(closedchan)
}

// ProgramContext returns a context that is canceled if one of the following
// signals are received by the program: os.Interrupt, syscall.SIGTERM,
// syscall.SIGQUIT. This is equivalent to passing a channel of size 1 to
// ContextWithCancelSignals that is notified by the same signals.
func ProgramContext() (context.Context, CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	return ContextWithCancelSignals(context.Background(), c)
}

// ContextWithCancelSignals returns a context that is cancelled when a signal
// is received on c (or if c is closed).
func ContextWithCancelSignals(parent context.Context, c <-chan os.Signal) (context.Context, CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	var mu sync.Mutex
	var s os.Signal

	// Lock signal retrival until we can confirm that s will no longer be
	// modified.
	mu.Lock()
	go func() {
		defer mu.Unlock()

		// Run until signal received or the (parent) context is canceled.
		select {
		case <-ctx.Done():
		case s = <-c:
			cancel()
		}
	}()

	// f will cancel ctx and return the value of s.
	f := func() os.Signal {
		cancel()
		mu.Lock()
		defer mu.Unlock()
		return s
	}

	return ctx, f
}

// ContextKey returns a concatinated operation key from context. Operation keys
// are added to context when an operation is started by a Handler. In the case
// of nested operations, keys are joined by a single dot (.). If there are no
// operation keys in context, an empty string is returned.
func ContextKey(ctx context.Context) string {
	s, _ := ctx.Value(opKey).(string)
	return s
}

func contextWithKey(ctx context.Context, key string) context.Context {
	if key == "" {
		return ctx
	}
	s := ContextKey(ctx)
	if s != "" {
		key = s + "." + key
	}

	return context.WithValue(ctx, opKey, key)
}
