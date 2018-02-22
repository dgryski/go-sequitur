package sequitur

import (
	"bytes"
	"fmt"
)

var testCompact = `Round and round the ragged rocks, the ragged rascal ran.`

func ExamplePrettyPrintCompact() {

	g := Parse([]byte(testCompact))

	var output bytes.Buffer
	if err := g.Compact().PrettyPrint(&output); err != nil {
		panic(err)
	}

	fmt.Println(string(output.Bytes()))

	// Output:
	// 1114369 -> {0 [R 1114374 a 1114371 r 1114374 1114385 o c k s ,   1114385 a s c a l 1114387 n .]}
	// 1114371 -> {2 [n 1114375]}
	// 1114374 -> {2 [o u 1114371]}
	// 1114375 -> {2 [d  ]}
	// 1114385 -> {2 [t h e 1114387 g g e 1114375 r]}
	// 1114387 -> {2 [  r a]}
}
