# CQRS

[![Go](https://github.com/GabrielCarpr/cqrs/actions/workflows/go.yml/badge.svg)](https://github.com/GabrielCarpr/cqrs/actions/workflows/go.yml)

CQRS is a WIP Go application framework centered around a command, query and event bus. Multiple interface adapters are available (JSON, HTML, CLI, GraphQL. Some are complete, some aren't) to connect to the bus. The interface adapters are intended to be generated from configuration files, or from the source itself. The bus can also be extended using middleware and plugins, for instance there's a background jobs manager included, and an event store.

A boilerplate application generator is also available:
`go run github.com/gabrielcarpr/cqrs/gen init [application-name] [root=.]`

An example of the generated application is in `_example`.

Other available commands:
`go run github.com/gabrielcarpr/cqrs/gen gen [rest/graphql]` Generates interface adapters
`go run github.com/gabrielcarpr/cqrs/gen make [command/query/test] [path] [name]` Generates a command, query, or test skeleto

## Project structure

- `_example` The generated boilerplate app as an example of a project
- `auth` A standalone auth package which contains a few utilities for auth and access control
- `background` A background jobs manager which extends the message bus
- `bus` The message bus
- `eventstore` An event store that can extend the message bus to store events (only a Postgres implementation is available for now)
- `gen` Command line utilties for generating code in a CQRS project
- `log` A very simple logging package
- `ports` The code for running interface adapters

Examples of how to use this library are in `_example`. More documentation will come once it's more complete and polished.
