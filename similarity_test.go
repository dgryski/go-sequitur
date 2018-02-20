package sequitur

import (
	"fmt"
)

func ExampleSimilarity() {

	texts := []string{testSimilarity, testImportance, testString, ""}
	textNames := []string{"sequitur.info", "wikipedia", "pease pudding", "empty"}

	cindex := make([]*CompactIndexed, len(texts))

	for textNum, text := range texts {
		cindex[textNum] = Parse([]byte(text)).Compact().Index(nil)
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
	// 0.00000   sequitur.info           empty
	// 0.05370       wikipedia   sequitur.info
	// 0.99648       wikipedia       wikipedia
	// 0.00289       wikipedia   pease pudding
	// 0.00000       wikipedia           empty
	// 0.00306   pease pudding   sequitur.info
	// 0.00289   pease pudding       wikipedia
	// 1.00000   pease pudding   pease pudding
	// 0.00000   pease pudding           empty
	// 0.00000           empty   sequitur.info
	// 0.00000           empty       wikipedia
	// 0.00000           empty   pease pudding
	// 0.00000           empty           empty

}

const testSimilarity = `
Sequitur is a method for inferring compositional hierarchies from strings. 
It detects repetition and factors it out of the string by forming rules in a grammar. 
The rules can be composed of non-terminals, giving rise to a hierarchy. 
It is useful for recognizing lexical structure in strings, and excels at very long sequences.

Craig Nevill-Manning, Google
Ian Witten, University of Waikato, New Zealand
` // http://www.sequitur.info/
