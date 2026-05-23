# tast

A TOML AST library for Go.

## Background

Built for [`protheon`](https://github.com/deahtstroke/protheon), which required making
surgical changes to small TOML configuration files at the AST level. The
existing TOML libraries for Go either lacked AST access entirely or exposed
unstable APIs that weren't suitable for production use, so this library was
built to cover that specific use case.

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
