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

// closedchan is a reusable closed channel.
var closedchan = make(chan struct{})

func init() {
	close(closedchan)
}

// SignalContext is a context.Context implementation that is canceled when the
// program recives a signal.
type SignalContext struct {
	context.Context

	mu     sync.Mutex
	done   chan struct{}
	signal os.Signal
}

// ProgramContext is a short-hand for ContextWithCancelSignals(
// context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT).
func ProgramContext() *SignalContext {
	return ContextWithCancelSignals(
		context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT,
	)
}

// ContextWithCancelSignals returns a context that listens for the passed in
// signals, and gets canceled once the first signal is received.
func ContextWithCancelSignals(parent context.Context, signals ...os.Signal) *SignalContext {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	parent, cancel := context.WithCancel(parent)
	ctx := &SignalContext{Context: parent}
	go func() {
		select {
		case <-parent.Done():
			ctx.mu.Lock()
			if ctx.done == nil {
				ctx.done = closedchan
			} else {
				close(ctx.done)
			}
			cancel() // Possibly redundant.
			ctx.mu.Unlock()
		case s := <-c:
			ctx.mu.Lock()
			ctx.signal = s
			if ctx.done == nil {
				ctx.done = closedchan
			} else {
				close(ctx.done)
			}
			cancel() // Needed to proagate cancel to children.
			ctx.mu.Unlock()
		}
	}()
	return ctx
}

// Signal returns nil if Done is not yet closed. If Done is closed due to a
// received signal, the received signal is returned.
func (ctx *SignalContext) Signal() os.Signal {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	return ctx.signal
}

// ExitCodeHint returns an exit code hint according to the signal in context and
// received error. The exit code follows UNIX conventions.
func (ctx *SignalContext) ExitCodeHint(err error) int {
	if err == nil {
		return 0
	}

	s := ctx.Signal()
	switch st := s.(type) {
	case syscall.Signal:
		return 128 + int(st)
	default:
		return 1
	}
}

// Done returns a channel that is closed when the context is canceled.
func (ctx *SignalContext) Done() <-chan struct{} {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if ctx.done == nil {
		ctx.done = make(chan struct{})
	}
	return ctx.done
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
