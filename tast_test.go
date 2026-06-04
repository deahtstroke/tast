package tast_test

import (
	"os"
	"testing"

	"github.com/deahtstroke/tast"
)

func TestParse(t *testing.T) {
	f, err := os.OpenFile("./testdata/test.toml", os.O_APPEND|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		t.Fatal("Unable to open file test.toml")
	}

	src, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Error reading from file test.toml")
	}

	_, err = tast.ParseBytes(src)
	if err != nil {
		t.Fatalf("Not expecting error, got: %v", err)
	}
}
