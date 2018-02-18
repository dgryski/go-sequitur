package sequitur

import (
	"fmt"
)

func ExampleSimilarity() {

	texts := []string{testSimilarity, testImportance, testString, " "}
	textNames := []string{"sequitur.info", "wikipedia", "pease pudding", "one space"}

	grammar := make([]*Grammar, len(texts))
	compact := make([]*Compact, len(texts))
	cindex := make([]*CompactIndexed, len(texts))

	for textNum, text := range texts {
		var err error

		grammar[textNum], err = Parse([]byte(text))
		if err != nil {
			panic(err)
		}

		compact[textNum] = grammar[textNum].Compact()

		cindex[textNum] = compact[textNum].Index(nil)
	}

	for i1, ci1 := range cindex {

		for i2, ci2 := range cindex {
			fmt.Printf("%7.5f %15s %15s\n", ci1.Similarity(ci2), textNames[i1], textNames[i2])
		}
	}

	// Output:
	// 1.00000   sequitur.info   sequitur.info
	// 0.05370   sequitur.info       wikipedia
	// 0.00306   sequitur.info   pease pudding
	// 0.00000   sequitur.info       one space
	// 0.05370       wikipedia   sequitur.info
	// 0.99648       wikipedia       wikipedia
	// 0.00289       wikipedia   pease pudding
	// 0.00000       wikipedia       one space
	// 0.00306   pease pudding   sequitur.info
	// 0.00289   pease pudding       wikipedia
	// 1.00000   pease pudding   pease pudding
	// 0.00000   pease pudding       one space
	// 0.00000       one space   sequitur.info
	// 0.00000       one space       wikipedia
	// 0.00000       one space   pease pudding
	// 1.00000       one space       one space

}

const testSimilarity = `
Sequitur is a method for inferring compositional hierarchies from strings. 
It detects repetition and factors it out of the string by forming rules in a grammar. 
The rules can be composed of non-terminals, giving rise to a hierarchy. 
It is useful for recognizing lexical structure in strings, and excels at very long sequences.

Craig Nevill-Manning, Google
Ian Witten, University of Waikato, New Zealand
` // http://www.sequitur.info/
