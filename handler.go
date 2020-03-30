package op

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"sync"
)

// HandlerError is the error type return from a Handler.
type HandlerError struct {
	ContextError    error
	OperationErrors map[string]error
}

func (err HandlerError) Error() string {
	n := len(err.OperationErrors)

	var sep, names string
	for k := range err.OperationErrors {
		names += sep + k
		sep = ", "
	}
	return fmt.Sprintf("%d operations failed: %v", n, names)
}

// Unwrap returns the handler's context error, if any.
func (err HandlerError) Unwrap() error {
	return err.ContextError
}

// Handler allows grouping and naming operatons.
type Handler struct {
	ctx context.Context
	wg  sync.WaitGroup
	mi  []MiddlewareFunc

	mu  sync.Mutex
	reg map[string]*Operation
}

// NewHandler returns a handler that will start operations using the passed in
// context. A sizeHint can optionally be provided.
func NewHandler(ctx context.Context, sizeHint ...int) *Handler {
	var hint int
	switch len(sizeHint) {
	case 0:
		hint = 64
	case 1:
		hint = sizeHint[0]
	default:
		panic("can only provide one sizeHint")
	}
	return &Handler{ctx: ctx, reg: make(map[string]*Operation, hint)}
}

// Use sets up a common middleware for use by all operatoins. It can be called
// multiple times to add multiple middleware functions. Middleware are applied
// in the oreder they are added, meaning the last middleware to be added will be
// the "outer" middleware.
func (h *Handler) Use(m MiddlewareFunc) {
	if m != nil {
		h.mi = append(h.mi, m)
	}
}

// Start adds the passed i operation to the handler, and starts it with an
// operator key appended to context. An incremental suffix is added to key if
// the key has previously been used within the same context.
func (h *Handler) Start(key string, op *Operation) {
	h.mu.Lock()
	key = uniqueKey(key, h.reg)
	h.reg[key] = op
	h.mu.Unlock()
	ctx := contextWithKey(h.ctx, key)

	// Apply common middelware.
	for _, m := range h.mi {
		op.Use(m)
	}
	op.Use(func(f Func) Func {
		return func(ctx context.Context) error {
			defer h.wg.Done()
			return f(ctx)
		}
	})

	// Start operation.
	h.wg.Add(1)
	op.Start(ctx)
}

var reIncrString = regexp.MustCompile(`_#(\d+)$`)

func uniqueKey(key string, m map[string]*Operation) string {
	_, invalid := m[key]
	invalid = invalid && key != ""
	if !invalid {
		return key
	}

	var base string
	var i int
	if loc := reIncrString.FindStringIndex(key); loc != nil {
		base = key[:loc[0]]
		i, _ = strconv.Atoi(key[loc[0]+2 : loc[1]])
	} else {
		base = key
	}

	for invalid {
		i++
		key = fmt.Sprintf("%s_#%d", base, i)
		_, invalid = m[key]
	}
	return key
}

// Wait waits for all started operations to completed, and returns an error if
// at least one of them failed.
func (h *Handler) Wait() error {
	h.wg.Wait()
	hErr := HandlerError{
		ContextError:    h.ctx.Err(),
		OperationErrors: make(map[string]error, len(h.reg)),
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	for k, op := range h.reg {
		err := op.Wait()
		if err != nil {
			hErr.OperationErrors[k] = err
		}
	}
	if len(hErr.OperationErrors) > 0 {
		return hErr
	}
	return nil
}
