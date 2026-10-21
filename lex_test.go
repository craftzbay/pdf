package pdf

import (
	"strings"
	"testing"
	"time"
)

// readObjectErr parses one object and reports whether the lexer rejected it.
// It fails the test if parsing does not finish promptly.
func readObjectErr(t *testing.T, input string, allowEOF bool) (rejected bool) {
	t.Helper()
	done := make(chan bool, 1)
	go func() {
		defer func() { done <- recover() != nil }()
		b := newBuffer(strings.NewReader(input), 0)
		b.allowEOF = allowEOF
		b.readObject()
	}()
	select {
	case rejected = <-done:
		return rejected
	case <-time.After(5 * time.Second):
		t.Fatalf("readObject(%q) did not return", input)
		return false
	}
}

func TestMalformedObjectsAreRejected(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    string
		allowEOF bool
	}{
		{"endobj in array", "1 0 obj [1 endobj", false},
		{"endobj in nested array", "1 0 obj << /K [[ endobj", false},
		{"array at EOF", "[1 2", true},
		{"hex string at EOF", "<0a", true},
		{"literal string at EOF", "(abc", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !readObjectErr(t, tc.input, tc.allowEOF) {
				t.Fatalf("readObject(%q) accepted malformed input", tc.input)
			}
		})
	}
}

func TestWellFormedObjectsStillParse(t *testing.T) {
	for _, input := range []string{
		"1 0 obj [1 2 0 R (a) <0a> /N] endobj",
		"1 0 obj endobj",
		"[1 2]",
		"(abc)",
		"<0a>",
	} {
		if readObjectErr(t, input, true) {
			t.Fatalf("readObject(%q) rejected well-formed input", input)
		}
	}
}
