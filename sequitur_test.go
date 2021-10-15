package sequitur

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/dgryski/go-tinyfuzz"
)

func TestNoInput(t *testing.T) {
	for _, in := range [][]byte{nil, []byte{}} {
		empty := Parse(in)
		var b bytes.Buffer
		if err := empty.Print(&b); err != nil {
			t.Error(err)
			return
		}
		if len(b.Bytes()) > 0 {
			t.Errorf("%#v input produced output: %v", in, b.Bytes())
		}
		var pp bytes.Buffer
		if err := empty.Print(&pp); err != nil {
			t.Error(err)
			return
		}
		if len(pp.String()) > 0 {
			t.Errorf("%#v input produced output: %v", in, pp.String())
		}
	}
}

func TestUTF8(t *testing.T) { // issue #3
	g := Parse([]byte(`°`))
	var buf bytes.Buffer
	g.PrettyPrint(&buf)
	if !utf8.Valid(buf.Bytes()) {
		t.Error("invalid utf8: " + buf.String())
	}
}

func TestPrintUTF8(t *testing.T) {
	g := Parse([]byte(testString))
	var buf bytes.Buffer
	g.Print(&buf)
	if buf.String() != testString {
		t.Error("UTF8 Print() incorrect\nWanted:\n"+testString,
			"Got:\n", buf.String())
	}
}

func TestPrintBinary(t *testing.T) {
	g := Parse(testBinary)
	var buf bytes.Buffer
	g.Print(&buf)
	if !bytes.Equal(buf.Bytes(), testBinary) {
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
		buf = append(buf[:0], "ab"...)
		buf = rb.appendBytes(buf)
		if want := "ab" + string([]rune{rune(i)}); string(buf) != want {
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
		if want := string([]rune{rune(i)}); got != want {
			t.Errorf("unexpected result, got %q want %q ", got, want)
		}
	}
}

func TestQuick(t *testing.T) {

	f := func(contents []byte) bool {
		if len(contents) == 0 {
			return true
		}

		var b bytes.Buffer

		g := Parse(contents)
		g.Print(&b)

		return bytes.Equal(b.Bytes(), contents)
	}

	if err := tinyfuzz.Fuzz(f, nil); err != nil {
		t.Errorf("error during quickcheck: %v", err)
	}
}

func TestGolden(t *testing.T) {

	corpusFiles, err := filepath.Glob("testdata/*.input")

	if err != nil {
		t.Errorf("error opening test_dir: %v", err)
		return
	}

	for _, corpusFile := range corpusFiles {
		contents, err := ioutil.ReadFile(corpusFile)
		if err != nil {
			t.Errorf("failed to read %s: %v", corpusFile, err)
			continue
		}

		var b bytes.Buffer

		g := Parse(contents)
		g.PrettyPrint(&b)

		outputFile := strings.TrimSuffix(corpusFile, ".input") + ".output"
		golden, err := ioutil.ReadFile(outputFile)
		if err != nil {
			t.Errorf("failed to read %s: %v", outputFile, err)
			continue
		}

		if !bytes.Equal(b.Bytes(), golden) {
			t.Errorf("mismatch for %s", corpusFile)
		} else {
			t.Logf("processed %s successfully", corpusFile)
		}

		b.Reset()
		g.Print(&b)
		if !bytes.Equal(b.Bytes(), contents) {
			t.Errorf("mismatch during evaluation for %s", corpusFile)
		} else {
			t.Logf("evaluated %s successfully", corpusFile)
		}
	}
}
