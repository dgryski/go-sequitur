package sequitur

import (
	"bytes"
	"reflect"
	"testing"
	"unicode/utf8"
)

func TestNoInput(t *testing.T) {
	g := new(Grammar)
	err := g.ParseBinary([]byte{})
	if err != ErrEmptyInput {
		t.Error("ErrEmptyInput not returned for empty input")
	}
}

func TestUTF8(t *testing.T) { // issue #3
	var g Grammar
	if err := g.ParseUTF8([]byte(`°`)); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	g.PrettyPrint(&buf)
	if !utf8.Valid(buf.Bytes()) {
		t.Error("invalid utf8: " + string(buf.Bytes()))
	}
}

func TestPrintUTF8(t *testing.T) {
	var g Grammar
	if err := g.ParseUTF8([]byte(testString)); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	g.Print(&buf)
	if string(buf.Bytes()) != testString {
		t.Error("UTF8 Print() incorrect\nWanted:\n"+testString,
			"Got:\n", string(buf.Bytes()))
	}
}

func TestPrintBinary(t *testing.T) {
	var g Grammar
	if err := g.ParseBinary(testBinary); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	g.Print(&buf)
	if !reflect.DeepEqual(buf.Bytes(), testBinary) {
		t.Error("Binary Print incorrect\nwanted:", testBinary, "\ngot:", buf.Bytes())
	}
}
