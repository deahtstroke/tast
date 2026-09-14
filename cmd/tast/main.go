package main

import (
	"os"

	"github.com/deahtstroke/tast/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
