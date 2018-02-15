package sequitur

import (
	"bytes"
	"fmt"
)

var testBinary = []byte{1, 2, 3, 4, 5, 'a', 'b', 5, 4, 3, 2, 1, 1, 'a', 'b', 2, 3, 4, 5, 2, 3}

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
	// 0 -> 0x01 1 2 0x05 0x04 0x03 0x02 0x01 0x01 2 1 3
	// 1 -> 3 0x04 0x05
	// 2 -> a b
	// 3 -> 0x02 0x03
}
