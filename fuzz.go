// +build gofuzz

package sequitur

import "bytes"

func Fuzz(data []byte) int {

	g := Parse(data)

	var b bytes.Buffer
	if err := g.Print(&b); err != nil {
		panic(err)
	}
	if !bytes.Equal(b.Bytes(), data) {
		panic("Parse/Print roundtrip mismatch")
	}

	gc := g.Compact()
	if !bytes.Equal(gc.Bytes(gc.RootID), data) {
		panic("Parse/Compact/Bytes roundtrip mismatch")
	}

	return 0
}
