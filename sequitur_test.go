package sequitur

import (
	"bytes"
	"reflect"
	"strconv"
	"testing"
	"unicode/utf8"
)

func TestNoInput(t *testing.T) {
	_, err := Parse([]byte{})
	if err != ErrEmptyInput {
		t.Error("ErrEmptyInput not returned for empty input")
	}
}

func TestUTF8(t *testing.T) { // issue #3
	g, err := Parse([]byte(`°`))
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
	g, err := Parse([]byte(testString))
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
	g, err := Parse(testBinary)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	g.Print(&buf)
	if !reflect.DeepEqual(buf.Bytes(), testBinary) {
		t.Error("Binary Print incorrect\nwanted:", testBinary, "\ngot:", buf.Bytes())
	}
}

func TestRuneOrByteAppendBytesWithByte(t *testing.T) {
	for i := 0; i <= 0xff; i++ {
		rb := newByte(byte(i))
		b := rb.appendBytes([]byte("ab"))
		if want := []byte{'a', 'b', byte(i)}; !bytes.Equal(b, want) {
			t.Errorf("unexpected bytes appended; got %q want %q", b, want)
		}
	}
}

func TestRuneOrByteAppendBytesWithRune(t *testing.T) {
	buf := make([]byte, 10)
	for i := 0; i <= utf8.MaxRune; i++ {
		rb := newRune(rune(i))
		buf := append(buf[:0], "ab"...)
		buf = rb.appendBytes(buf)
		if want := "ab" + string(i); string(buf) != want {
			t.Errorf("unexpected bytes appended; got %q want %q", buf, want)
		}
	}
}

func TestRuneOrByteAppendEscapedWithByte(t *testing.T) {
	buf := make([]byte, 10)
	for i := 0; i <= 0xff; i++ {
		if i == '"' || i == '\\' {
			continue
		}
		buf = append(buf[:0], '"')
		rb := newByte(byte(i))
		buf = rb.appendEscaped(buf)
		buf = append(buf, '"')
		got, err := strconv.Unquote(string(buf))
		if err != nil {
			t.Errorf("cannot unquote %q (byte %x)", buf, i)
			continue
		}
		if want := []byte{byte(i)}; !bytes.Equal([]byte(got), want) {
			t.Errorf("unexpected result, got %q want %q ", got, want)
		}
	}
}

func TestRuneOrByteAppendEscapedWithRune(t *testing.T) {
	buf := make([]byte, 20)
	for i := 0; i <= utf8.MaxRune; i++ {
		if i == '"' || i == '\\' {
			continue
		}
		buf = append(buf[:0], '"')
		rb := newRune(rune(i))
		buf = rb.appendEscaped(buf)
		buf = append(buf, '"')
		got, err := strconv.Unquote(string(buf))
		if err != nil {
			t.Fatalf("cannot unquote %q (byte %x): %v", buf, i, err)
			continue
		}
		if want := string(i); got != want {
			t.Fatalf("unexpected result, got %q want %q ", got, want)
		}
	}
}
