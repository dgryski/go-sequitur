package sequitur

import (
	"bytes"
	"reflect"
	"testing"
	"unicode/utf8"
)

func TestNoInput(t *testing.T) {
	_, err := ParseBinary([]byte{})
	if err != ErrEmptyInput {
		t.Error("ErrEmptyInput not returned for empty input")
	}
}

func TestUTF8(t *testing.T) { // issue #3
	g, err := ParseUTF8([]byte(`°`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	g.PrettyPrint(&buf)
	if !utf8.Valid(buf.Bytes()) {
		t.Error("invalid utf8: " + string(buf.Bytes()))
	}
}

func TestPrintUTF8(t *testing.T) {
	g, err := ParseUTF8([]byte(testString))
	if err != nil {
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
	g, err := ParseBinary(testBinary)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	g.Print(&buf)
	if !reflect.DeepEqual(buf.Bytes(), testBinary) {
		t.Error("Binary Print incorrect\nwanted:", testBinary, "\ngot:", buf.Bytes())
	}
}
