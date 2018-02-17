package sequitur

import (
	"fmt"
)

func ExampleImportance() {

	g, err := Parse([]byte(testImportance))
	if err != nil {
		panic(err)
	}

	for k, v := range g.Compact().Index(5, true).Importance() {

		if k >= 10 {
			break
		}

		fmt.Println(k, v.Score, v.Used, string(v.Bytes))

	}

	// Output:
	// 0 225 5 algorithm
	// 1 176 2 Nevill-Manning, C.G.; Witten, I.H. (1997). "
	// 2 175 5 grammar
	// 3 162 3 nonterminal symbol
	// 4 128 4 sequence
	// 5 126 3 in the grammar
	// 6 96 4 in the
	// 7 96 4 digram
	// 8 96 4 symbol
	// 9 96 4 equenc
}

const testImportance = `
Sequitur algorithm
From Wikipedia, the free encyclopedia
Sequitur (or Nevill-Manning algorithm) is a recursive algorithm developed by Craig Nevill-Manning and Ian H. Witten in 1997[1] that infers a hierarchical structure (context-free grammar) from a sequence of discrete symbols. The algorithm operates in linear space and time. It can be used in data compression software applications.[2]

Contents 
1	Constraints
1.1	Digram uniqueness
1.2	Rule utility
2	Method summary
3	See also
4	References
5	External links
Constraints
The sequitur algorithm constructs a grammar by substituting repeating phrases in the given sequence with new rules and therefore produces a concise representation of the sequence. For example, if the sequence is

S→abcab,
the algorithm will produce

S→AcA, A→ab.
While scanning the input sequence, the algorithm follows two constraints for generating its grammar efficiently: digram uniqueness and rule utility.

Digram uniqueness
Whenever a new symbol is scanned from the sequence, it is appended with the last scanned symbol to form a new digram. If this digram has been formed earlier then a new rule is made to replace both occurrences of the digrams. Therefore, it ensures that no digram occurs more than once in the grammar. For example, in the sequence S→abaaba, when the first four symbols are already scanned, digrams formed are ab, ba, aa. When the fifth symbol is read, a new digram 'ab' is formed which exists already. Therefore, both instances of 'ab' are replaced by a new rule (say, A) in S. Now, the grammar becomes S→AaAa, A→ab, and the process continues until no repeated digram exists in the grammar.

Rule utility
This constraint ensures that all the rules are used more than once in the right sides of all the productions of the grammar, i.e., if a rule occurs just once, it should be removed from the grammar and its occurrence should be substituted with the symbols from which it is created. For example, in the above example, if one scans the last symbol and applies digram uniqueness for 'Aa', then the grammar will produce: S→BB, A→ab, B→Aa. Now, rule 'A' occurs only once in the grammar in B→Aa. Therefore, A is deleted and finally the grammar becomes

S→BB, B→aba.
This constraint helps reduce the number of rules in the grammar.

Method summary
The algorithm works by scanning a sequence of terminal symbols and building a list of all the symbol pairs which it has read. Whenever a second occurrence of a pair is discovered, the two occurrences are replaced in the sequence by an invented nonterminal symbol, the list of symbol pairs is adjusted to match the new sequence, and scanning continues. If a pair's nonterminal symbol is used only in the just created symbol's definition, the used symbol is replaced by its definition and the symbol is removed from the defined nonterminal symbols. Once the scanning has been completed, the transformed sequence can be interpreted as the top-level rule in a grammar for the original sequence. The rule definitions for the nonterminal symbols which it contains can be found in the list of symbol pairs. Those rule definitions may themselves contain additional nonterminal symbols whose rule definitions can also be read from elsewhere in the list of symbol pairs.[3]

See also
Context-free grammar
Data compression
Lossless data compression
Straight-line grammar
References
 Nevill-Manning, C.G.; Witten, I.H. (1997). "Identifying Hierarchical Structure in Sequences: A linear-time algorithm". arXiv:cs/9709102 Freely accessible.
 Nevill-Manning, C.G.; Witten, I.H. (1997). "Linear-Time, Incremental Hierarchy Inference for Compression". doi:10.1109/DCC.1997.581951.
 GrammarViz 2.0 – Sequitur and parallel Sequitur implementations in Java, Sequitur-based time series patterns discovery
External links
sequitur.info – the reference Sequitur algorithm implementation in C++, Java, and other languages
`
