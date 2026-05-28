# tast

A TOML AST library for Go.

## Background

Built for [`protheon`](https://github.com/deahtstroke/protheon), which required making
surgical changes to small TOML configuration files at the AST level. The
existing TOML libraries for Go either lacked AST access entirely or exposed
unstable APIs that weren't suitable for production use, so this library was
built to cover that specific use case. The current implementation of the AST
itself does not support using Go's standard interfaces for readers and writers, namely
`io.Reader` and `io.Writer`, since the scope of this project is focused mostly on very small
TOML configurations that can be parsed in one go without relying on internal buffering
and look-ahead logic.

> [!WARNING]
> The library is as of time of writing not 100% compliant with TOML spec v1.1.0
>
> Use at your own discretion

## Installation

    go get github.com/deahtstroke/tast

## Usage

```go
doc, err := tast.Parse(src)
if err != nil {
    // handle *tast.ParseError
}

doc.WriteFile("config.toml")
```

## Status

Early development — API may shift while on `v0.x.x`.

