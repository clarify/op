# Operation [![GitHub Actions](https://github.com/searis/template-go/workflows/Go/badge.svg?branch=master)](https://github.com/searis/template-go/actions?query=workflow%3AGo+branch%3Amaster) [![GoDev](https://img.shields.io/static/v1?label=go.dev&message=reference&color=blue)](https://pkg.go.dev/github.com/searis/op)

The Operation package provides utilities for controlling go-routines and the program run-time by use of [context](https://golang.org/pkg/context). It is built-up by the following main concepts:

- `SignalContext`: Listen for signals to cancel context, retrieve exit code hints.
- `Operations`: Manage/wait for a single go-routine.
- `Handler`: Manage multiple named operations, insert operation key(s) to context.
- `MiddlewareFunc`: Can be passed to a single Operation or to a Handler.

See [Full Application example](/examples/full-program/main.go).
