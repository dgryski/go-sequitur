package sequitur

import (
	"bytes"
	"fmt"
)

var testBinary = []byte{0xfe, 0xff, 0xfd, 0xfe, 0xff, 1, 2, 3, 4, 5, 'a', 'b', 1, 2, 3, 4, 5, 'a', 'b'}

func ExamplePrettyPrintBinary() {

	g := new(Grammar)

	err := g.ParseBinary(testBinary)
	if err != nil {
		panic(err)
	}

	var output bytes.Buffer
	err = g.PrettyPrint(&output)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(output.Bytes()))

	// Output:
	// 0 -> 1 0xFD 1 2 2
	// 1 -> 0xFE 0xFF
	// 2 -> 0x01 0x02 0x03 0x04 0x05 a b
}
