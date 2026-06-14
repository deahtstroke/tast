# tast

A TOML AST library for Go.

## Background

Built for [`protheon`](https://github.com/deahtstroke/protheon), which required making
surgical changes to small TOML configuration files at the AST-level. The
existing TOML libraries for Go either lacked AST access entirely or exposed
unstable APIs that weren't suitable for production use, so this library was
built to cover that specific use case. The current implementation of the AST
itself does not support using Go's standard interfaces for readers and writers,
namely `io.Reader` and `io.Writer`, since the scope of this project is focused
mostly on very small TOML configurations that can be parsed in one go without
relying on internal buffering and look-ahead logic.

## Who is Tast for?

The big purpose for making Tast in the first place, was to get the benefit of what
an AST gives intrinsically: The ability to surgically edit the source tree of a file.
Therefore, people that need to edit TOMLs files but want to preserve trivia/comments,
like was my case, would find using this library beneficial

## Installation

    go get github.com/deahtstroke/tast

## Usage

Given the following TOML file `db_config.toml`

```toml
# My database config
[database]
username = "root_username"
password = "root_password"
port = 5432
host = "localhost"
```

Read the TOML file as a document in your Go project

```go
import (
    "github.com/deahtstroke/tast"
    "os"
)

const (
    path = "path/to/some/file"
)

func main() {
    src, _ := os.ReadFile(path)
    doc, err := tast.ParseBytes(src)
    if err != nil {
        // handle tast errors
    }

    dbTable, ok := doc.Table("database")
    if !ok {
        // ...
    }

    if err = dbTable.Set("username", "new_username"); err != nil {
        // ...
    }

    // finally, save back to the original file
    err = doc.Save("db_config.toml")
    if err != nil {
        // ...
    }
}
```

The result for this small mutation in the `db_config.toml` file would be:

```toml
# My database config
[database]
username = "new_username" # <-- Mutation occurred here!
password = "root_password"
port = 5432
host = "localhost"

```

Ordering, comments, and formatting is preserved, which is exactly what
Tast aims to do.

## Status

Early development — API may shift while on `v0.x.x`.
