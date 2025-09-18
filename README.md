# Ark REPL

[![Test status](https://img.shields.io/github/actions/workflow/status/mlange-42/ark-repl/tests.yml?branch=main&label=Tests&logo=github)](https://github.com/mlange-42/ark-repl/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mlange-42/ark-repl)](https://goreportcard.com/report/github.com/mlange-42/ark-repl)
[![Go Reference](https://pkg.go.dev/badge/github.com/mlange-42/ark-repl.svg)](https://pkg.go.dev/github.com/mlange-42/ark-repl)
[![GitHub](https://img.shields.io/badge/github-repo-blue?logo=github)](https://github.com/mlange-42/ark-repl)
[![MIT license](https://img.shields.io/badge/MIT-brightgreen?label=license)](https://github.com/mlange-42/ark-repl/blob/main/LICENSE-MIT)
[![Apache 2.0 license](https://img.shields.io/badge/Apache%202.0-brightgreen?label=license)](https://github.com/mlange-42/ark-repl/blob/main/LICENSE-APACHE)

*Ark REPL* provides a REPL for inspecting applications made with the [Ark](https://github.com/mlange-42/ark) Entity Component System (ECS).

<div align="center">

<a href="https://github.com/mlange-42/ark">
<img src="https://github.com/user-attachments/assets/4bbe57c6-2e16-43be-ad5e-0cf26c220f21" alt="Ark (logo)" width="500px" />
</a>

</div>

## Features

- Interactive inspection of World internals.
- Command help inside the REPL.
- Can control the update loop (pause, resume, stop).
- Can connect from a separate terminal.
- Extensible: add your own commands.

## Installation

Add the library to your Ark application:

```
go get github.com/mlange-42/ark-repl
```

Install CLI for a REPL in a separate terminal:

```
go install github.com/mlange-42/ark-repl/cmd/ark
```

## Usage

See the [API docs](https://pkg.go.dev/github.com/mlange-42/ark-repl) and [examples](https://github.com/mlange-42/ark-repl/tree/main/examples) for library usage.

For starting the REPL in a separate terminal, run this after starting your Ark application:

```
ark
```

## License

This project is distributed under the [MIT license](./LICENSE-MIT) and the [Apache 2.0 license](./LICENSE-APACHE), as your options.