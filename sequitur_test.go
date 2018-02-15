package sequitur

import (
	"bytes"
	"testing"
	"unicode/utf8"
)

func TestNoInput(t *testing.T) {
	g := new(Grammar)
	err := g.Parse([]byte{})
	if err != ErrEmptyInput {
		t.Error("ErrEmptyInput not returned for empty input")
	}
}

func TestUTF8(t *testing.T) { // issue #3
	var g Grammar
	if err := g.Parse([]byte(`°`)); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	g.PrettyPrint(&buf)
	if !utf8.Valid(buf.Bytes()) {
		t.Error("invalid utf8: " + string(buf.Bytes()))
	}
}

func TestPrint(t *testing.T) {
	var g Grammar
	if err := g.Parse([]byte(testString)); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	g.Print(&buf)
	if string(buf.Bytes()) != testString {
		t.Error("Print() incorrect\nWanted:\n"+testString,
			"Got:\n", string(buf.Bytes()))
	}
}
