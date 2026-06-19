package tast

import (
	"io"
	"os"
)

// Reads the a TOML document from a path to a source file
func LoadFile(path string) (*Document, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBytes(f)
}

// Parses a byte array into a TOML document
func ParseBytes(src []byte) (*Document, error) {
	scanner := NewScanner(src)
	scanner.Scan()

	parser := NewParser(scanner.tokens)
	doc, errs := parser.parse()
	if len(errs) > 0 {
		return nil, errs[0]
	}
	return doc, nil
}

// Parses a string into a TOML document
func ParseString(src string) (*Document, error) {
	return ParseBytes([]byte(src))
}

func ParseFrom(r io.Reader) (*Document, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return ParseBytes(src)
}

// Saves the current document source to a file
func (d *Document) Save(path string) error {
	s, err := NewPrinter().print(d)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(s), 0o644)
}
