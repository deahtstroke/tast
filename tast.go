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

func ParseBytes(src []byte) (*Document, error) {
	scanner := NewScanner(src)
	scanner.Scan()

	parser := NewParser(scanner.Tokens)
	doc, errs := parser.parse()
	if len(errs) > 0 {
		return nil, errs[0]
	}
	return doc, nil
}

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

func (d *Document) Save(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	s, err := NewPrinter().print(d)
	if err != nil {
		return err
	}

	_, err = io.WriteString(f, s)
	return err
}
