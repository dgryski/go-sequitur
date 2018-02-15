package sequitur

import (
	"testing"
)

func TestNoInput(t *testing.T) {
	g := new(Grammar)
	err := g.Parse([]byte{})
	if err != ErrEmptyInput {
		t.Error("ErrEmptyInput not returned for empty input")
	}
}
