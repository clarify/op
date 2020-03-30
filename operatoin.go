package op

import (
	"context"
)

// Func describes an operation function.
type Func func(ctx context.Context) error

// Operation manages the runtime of a go-routine, and is intended to be run
// exactly once. Once started, it can be either waited for or canceled.
type Operation struct {
	f      Func
	cancel func()
	done   chan struct{}

	err error
}

// New returns a new operation for f.
func New(f Func) *Operation {
	return &Operation{
		f: f,
	}
}

// Use applies a middleware to the operation. This must be done before the
// operation is started. It can be called multiple times to add multiple
// middleware functions. Middleware are applied in the oreder they are added,
// meaning the last middleware to be added will be the "outer" middleware.
func (op *Operation) Use(m MiddlewareFunc) {
	op.f = m(op.f)
}

// Start the operation in the background using the passed in context. This
// method must be called before you can wait for or cancel the operation.
// Concurrent calls to Start are invalid, and will result in unspecified
// behavior.
func (op *Operation) Start(ctx context.Context) {
	if op.done != nil {
		return
	}
	op.done = make(chan struct{})

	ctx, cancel := context.WithCancel(ctx)
	op.cancel = cancel

	go func() {
		op.err = op.f(ctx)
		cancel()
		close(op.done)
	}()
}

// Wait waits for the operation to be complete, and return an error when the
// operation returned an error. It is invalid to call this method before the
// operation is started.
func (op *Operation) Wait() error {
	if op.done == nil {
		panic("operation not started")
	}
	<-op.done
	return op.err
}

// Cancel cancels the context passed to the operation. It is invalid to call
// this method before the operation is started.
func (op *Operation) Cancel() {
	if op.done == nil {
		panic("operation not started")
	}
	op.cancel()
}
