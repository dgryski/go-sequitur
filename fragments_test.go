package sequitur

import (
	"bytes"
	"fmt"
	"testing"
)

func TestFragments(t *testing.T) {

	testSlices := [][]byte{}
	testSlices = append(testSlices, []byte(testString))
	testSlices = append(testSlices, testBinary)
	testSlices = append(testSlices, []byte(testCompact))

	for tNum, test := range testSlices {
		g := Parse(test)
		if string(g.Symbol().Bytes()) != string(test) {
			t.Error(tNum, "g.Symbol().Bytes() not equal")
		}

		comp := g.Compact()
		if string(comp.Bytes(comp.RootID)) != string(test) {
			t.Error(tNum, "comp.Bytes() not equal")
		}
	}
}

func TestEmpty(t *testing.T) {
	g := Grammar{}
	if len(g.Symbol().Bytes()) > 0 {
		t.Error("Empty Grammar returns non-empty Bytes():", g.Symbol().Bytes())
	}
	if g.Symbol().String() != EmptySymbolIDstring {
		t.Error("Empty Grammar does not return '"+EmptySymbolIDstring+"' from String():", g.Symbol().String())
	}
	if len(g.Symbol().SubSymbols()) > 0 {
		t.Error("Empty Grammar returns non-empty SubSymbols():", g.Symbol().SubSymbols())
	}
	c := g.Compact()
	if len(c.Bytes(c.RootID)) > 0 {
		t.Error("Empty CompactGrammar returns non-empty Bytes():", c.Bytes(c.RootID))
	}
	if fmt.Sprintf("%v", c) != EmptySymbolIDstring {
		t.Error("Empty CompactGrammar does not return '"+EmptySymbolIDstring+"' from String():", c.String())
	}
	var b bytes.Buffer
	err := c.PrettyPrint(&b)
	if err != nil {
		panic(err)
	}
	if len(b.String()) > 0 {
		t.Error("Empty CompactGrammar returns non-empty PrettyPrint():", b.String())
	}

}
