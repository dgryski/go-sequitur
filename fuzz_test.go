package sequitur

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestFuzz(t *testing.T) {
	corpusFiles, err := filepath.Glob("workdir/corpus/*")

	if err != nil {
		t.Skipf("error opening test_dir: %v", err)
		return
	}

	for _, corpusFile := range corpusFiles {
		contents, err := ioutil.ReadFile(corpusFile)

		if err != nil {
			t.Errorf("failed to read %s: %v", corpusFile, err)
			continue
		}

		if len(contents) == 0 {
			continue
		}

		var b bytes.Buffer

		g := Parse(contents)
		g.Print(&b)
		if !bytes.Equal(b.Bytes(), contents) {
			t.Errorf("mismatch during evaluation for %s", corpusFile)
		} else {
			t.Logf("evaluated %s successfully", corpusFile)
		}
	}
}
