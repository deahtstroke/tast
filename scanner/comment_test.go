package scanner

import "testing"

func Test_CommentNode(t *testing.T) {
	str := `# Some comment here
# Some comment there`

	scanner := Scanner{
		Source:  []byte(str),
		line:    0,
		current: 0,
		start:   0,
	}
	scanner.Scan()

	if len(scanner.Tokens)-1 != 2 { // Minus the EOF token
		t.Fatalf("Expecting two comment tokens parsed: %v", len(scanner.Tokens))
	}
}
