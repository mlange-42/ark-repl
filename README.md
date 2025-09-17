# Ark Inspect

[![Test status](https://img.shields.io/github/actions/workflow/status/mlange-42/ark-inspect/tests.yml?branch=main&label=Tests&logo=github)](https://github.com/mlange-42/ark-inspect/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mlange-42/ark-inspect)](https://goreportcard.com/report/github.com/mlange-42/ark-inspect)
[![Go Reference](https://pkg.go.dev/badge/github.com/mlange-42/ark-inspect.svg)](https://pkg.go.dev/github.com/mlange-42/ark-inspect)
[![GitHub](https://img.shields.io/badge/github-repo-blue?logo=github)](https://github.com/mlange-42/ark-inspect)
[![MIT license](https://img.shields.io/badge/MIT-brightgreen?label=license)](https://github.com/mlange-42/ark-inspect/blob/main/LICENSE-MIT)
[![Apache 2.0 license](https://img.shields.io/badge/Apache%202.0-brightgreen?label=license)](https://github.com/mlange-42/ark-inspect/blob/main/LICENSE-APACHE)

*Ark Inspect* provides a REPL for inspection of running applications using the [Ark](https://github.com/mlange-42/ark) Entity Component System (ECS).

<div align="center">

<a href="https://github.com/mlange-42/ark">
<img src="https://github.com/user-attachments/assets/4bbe57c6-2e16-43be-ad5e-0cf26c220f21" alt="Ark (logo)" width="500px" />
</a>

</div>

## Features

* Scheduler for running logic and UI systems with independent update rates.
* Interfaces for ECS systems and observers.
* Ready-to-use systems for common tasks like writing CSV files or terminating a simulation.
* Common ECS resources, like central PRNG source or the current update tick.

## Installation

### Library

```
go get github.com/mlange-42/ark-inspect
```

### CLI for using from a second terminal

```
go install github.com/mlange-42/ark-inspect/cmd/ark
```

## Usage

See the [API docs](https://pkg.go.dev/github.com/mlange-42/ark) and [examples](https://github.com/mlange-42/ark/tree/main/examples) for details.

## License

This project is distributed under the [MIT license](./LICENSE-MIT) and the [Apache 2.0 license](./LICENSE-APACHE), as your options.