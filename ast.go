package tast

import (
	"io"
	"os"

	"github.com/deahtstroke/toml-ast/scanner"
)

func Parse(src []byte) (*Document, error) {
	scanner := scanner.NewScanner(src)
	scanner.Scan()

	parser := NewParser(scanner.Tokens)
	doc, errs := parser.parse()
	if len(errs) > 0 {
		return nil, errs[0]
	}
	return doc, nil
}

func ParseString(src string) (*Document, error) {
	return Parse([]byte(src))
}

func ParseFile(path string) (*Document, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(file)
}

func (d *Document) Format() (string, error) {
	printer := NewPrinter()
	return printer.print(d)
}

func (d *Document) Write(w io.Writer) error {
	s, err := d.Format()
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

func (d *Document) WriteFile(path string) error {
	s, err := d.Format()
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(s), 0o644)
}
