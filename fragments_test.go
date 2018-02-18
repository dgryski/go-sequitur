package sequitur

import (
	"reflect"
	"testing"
)

func TestFragments(t *testing.T) {

	testSlices := [][]byte{}
	testSlices = append(testSlices, []byte(testString))
	testSlices = append(testSlices, testBinary)
	testSlices = append(testSlices, []byte(testCompact))

	for tNum, test := range testSlices {
		g, err := Parse(test)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(g.Symbol().Bytes(), test) {
			t.Error(tNum, "g.Symbol().Bytes() not equal")
		}

		comp := g.Compact()
		if !reflect.DeepEqual(comp.Bytes(comp.RootID), test) {
			t.Error(tNum, "comp.Bytes() not equal")
		}

	}
}
