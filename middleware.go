package op

import "context"

// MiddlewareFunc describes a closure thar return a middleware for the passed in
// function. A MiddelwareFunc can be added to an Operation or Handler.
type MiddlewareFunc func(Func) Func

// OnError returns a middleware that issues the passed in callback function
// when the inner function fails.
func OnError(callback func()) MiddlewareFunc {
	return func(f Func) Func {
		return func(ctx context.Context) error {
			err := f(ctx)
			if err != nil {
				callback()
			}
			return err
		}
	}
}
