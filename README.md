# Operation [![GitHub Actions](https://github.com/clarify/op/workflows/Go/badge.svg?branch=master)](https://github.com/clarify/op/actions?query=workflow%3AGo+branch%3Amaster) [![GoDev](https://img.shields.io/static/v1?label=go.dev&message=reference&color=blue)](https://pkg.go.dev/github.com/clarify/op)

The Operation package provides utilities for controlling go-routines and the program run-time by use of [context](https://golang.org/pkg/context). It is built-up by the following main concepts.

## ProgramContext and exit codes

The `op.ProgramContext` function allows initialing a context that is cancelled when a termination signal is retrieved by the program. The returned cancel function can be used to retrieve the termination signal, if one was received, so that it can be passed along to the `op.ExitCodeHint` function alongside an error.

## Operations

The `op.Operation` type can be used start, cancel and wait for a single go-routine to complete. They are initialized by passing in a `Func` instance, which is a function accepting a context and returning an error. While a `Func` instance can be reused, an operation can _only_ be started once. Cleaver use of operations allows you to express dependencies and more.

An operation can also be wrapped by reusable middle-ware before it's starts by calling the `Use` method. This can be useful for managing e.g. logging, tracing or errors. When it's useful, middle-ware can also be called directly against a `Func` instance. A possible use-case for this would be if the func instance is reused to start multiple operations.

## Handler

The `op.Handler` type (read "operation handler") allow managing multiple operations through use of a common base context.

Each operation started by a handler will receive a key in it's context that is unique to the handler instance. If multiple handler instances are nested, this key will form a dot-joined _path_. The key can be useful for logging or tracing of operation.

A handler also allow middle-ware to be added. By doing so, you can describe a common set of middle-ware for all operations that are to be started by a given handler. Be awere that handler middelware will always wrap operation middelware.


## Middle-ware

Functions that implement the `op.MiddlewareFunc` interface can be called to directly wrap `op.Func` instances, or they can be passed to `op.Handler` or `op.Operation` instances by passing them to the instance `Use` method. Middle-ware are always applied "on top", so that the last call to the `Use` method describes the _outer_ method. Middle-ware added to a handler instance will wrap any middle-ware added directly to the Operation instance, just like middle-ware added to an operation instance will wrap any middle-ware applied to a Func instance.

To understand which order middle-ware run, it might help to look on middle-ware as layers of an onion. Every time you call the `Use` method on an operation, you will add another outer layer. When starting an operation through a handler, you will add the full set of layers already added to the handler. The core of the onion is the inner `Func` instance. When starting the operation, we stick a knitting-pin into the onion trying to reach it's core. Each layer may allow us to continue the penetration, add it's flavours to the tip, or force us to abort. Once we either reach the core or a layer that is impenetrable, we drag it back out, passing trough all the same layers in opposite order, which again adds (or remove) flavors to the tip.

PS! For cutting onions in real life, use a knife!

## Full example

To see all these concepts in practice, navigate to the [full program example](/examples/full-program/main.go).
