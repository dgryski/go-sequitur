package sequitur

import (
	"bytes"
	"fmt"
)

const peasePudding = `
pease porridge hot,
pease porridge cold,
pease porridge in the pot,
nine days old.

some like it hot,
some like it cold,
some like it in the pot,
nine days old.
`

func ExamplePrettyPrint() {

	g := new(Grammar)

	err := g.Parse([]byte(peasePudding))
	if err != nil {
		panic(err)
	}

	var output bytes.Buffer
	err = g.PrettyPrint(&output)
	if err != nil {
		panic(err)
	}

	// this next line just required to make the output correct for the test to pass
	fmt.Println(string(bytes.Replace(output.Bytes(), []byte(" \n"), []byte("\n"), -1)))

	// Output:
	// 0 -> 1 2 3 4 3 5 \n 6 2 7 4 7 5
	// 1 -> \n p e a s 8 r r i d g 9
	// 2 -> h o t
	// 3 -> , 1
	// 4 -> c 10
	// 5 -> 11 _ t h 8 t 12 n 11 9 d a y s _ 10 . \n
	// 6 -> s o m 9 l i k 9 i t _
	// 7 -> 12 6
	// 8 -> 9 p o
	// 9 -> e _
	// 10 -> o l d
	// 11 -> i n
	// 12 -> , \n
}
