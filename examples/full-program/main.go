package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/clarify/op"
)

func main() {
	// Parse program configuration or exit with exit code 2.
	p := mustParseProgram(os.Args[1:])

	// Get a context that is automatically canceled when the program receives a
	// termination signal.
	ctx, cancel := op.ProgramContext()

	// Run program and handle errors.
	err := p.run(ctx)
	sig := cancel()
	switch {
	case err == nil:
		log.Printf("I! Program succeed")
	case sig != nil:
		log.Printf("E! Program aborted by user: %v", err)
	default:
		log.Printf("E! Program internal failure: %v", err)
	}

	os.Exit(op.ExitHint(sig, err))
}

type program struct {
	foo bool
}

func mustParseProgram(args []string) program {
	var p program
	set := flag.NewFlagSet("my-program", flag.ExitOnError)
	set.BoolVar(&p.foo, "foo", false, "Enable foo operation")
	_ = set.Parse(args)
	return p
}

func (p program) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	oh := op.NewHandler(ctx)

	// Set up common middleware to apply to all operations before start.
	oh.Use(op.OnError(cancel)) // Cancel all operations on first error.
	oh.Use(withLog)

	// Define operations to run. Each operation can be run at most once.
	op1 := op.New(simulateWork(1 * time.Second))
	op2 := op.New(simulateWork(2 * time.Second))
	foo := op.New(simulateNonInterruptibleWork(3 * time.Second))

	// Individual operations can get middleware as well.
	foo.Use(func(f op.Func) op.Func {
		return func(ctx context.Context) error {
			k := op.ContextKey(ctx)
			log.Printf("D! %s: Hello from foo middleware!", k)
			return f(ctx)
		}
	})

	oh.Start("op1", op1)
	oh.Start("op2", op2)
	if p.foo {
		// We can wait for individual operations to complete; in this example
		// we have set "foo" to only start if op1 completes without error.
		err := op1.Wait()
		if err == nil {
			oh.Start("foo", foo)
		}
	}

	// Wait for all operations to complete.
	return oh.Wait()
}

// withLog shows an example middleware function that logs operation events using
// the standard library logger.
func withLog(f op.Func) op.Func {
	return func(ctx context.Context) error {
		start := time.Now()

		k := op.ContextKey(ctx)
		log.Printf("D! %s: Operation started", k)
		err := f(ctx)

		d := time.Since(start)
		switch {
		case err == nil:
			log.Printf("D! %s: Operation completed in %v", k, d)
		case errors.Is(err, context.Canceled):
			log.Printf("I! %s: Operation canceled after %v: %v", k, d, err)
		case errors.Is(err, context.DeadlineExceeded):
			log.Printf("I! %s: Operation timed out after %v: %v", k, d, err)
		default:
			log.Printf("E! %s: Operation failed after %v: %v", k, d, err)
		}

		return err
	}
}

func simulateWork(d time.Duration) op.Func {
	return func(ctx context.Context) error {
		// simulating interruptible work.
		select {
		case <-time.After(d):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func simulateNonInterruptibleWork(d time.Duration) op.Func {
	return func(ctx context.Context) error {
		c := time.After(d)
		select {
		case <-c:
			return nil
		case <-ctx.Done():
			k := op.ContextKey(ctx)
			log.Printf("D! %s: Ignoring cancellation attempt", k)
		}
		<-c
		return nil
	}
}
