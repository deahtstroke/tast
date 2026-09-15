package tast

import (
	"io"
	"os"
)

// Parses a byte array into a TOML document
func ParseBytes(src []byte) (*Document, error) {
	tokens, err := newScanner(src).scan()
	if err != nil {
		return nil, err
	}

	parser := newParser(tokens)
	doc, errs := parser.parse()
	if len(errs) > 0 {
		return nil, errs[0]
	}
	return doc, nil
}

// Reads the TOML document from a path to a source file
func LoadFile(path string) (*Document, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBytes(f)
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

func (d *Document) WriteTo(w io.Writer) (int64, error) {
	p := newPrinter(w)
	if err := p.print(d); err != nil {
		return 0, p.err
	}

	return p.n, p.err
}

// Saves the current document source to a file
func (d *Document) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = d.WriteTo(f)
	return err
}
