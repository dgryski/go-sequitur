// Package sequitur implements the sequitur algorithm
/*
	http://www.sequitur.info/
	https://en.wikipedia.org/wiki/Sequitur_algorithm

*/
package sequitur

import (
	"errors"
	"fmt"
	"io"
)

// Grammar is a constructed grammar.  The zero value is safe to call Parse on.
type Grammar struct {
	table  digrams
	base   *rules
	ruleID uint64
}

func (g *Grammar) nextID() uint64 {
	g.ruleID++
	return g.ruleID
}

type rules struct {
	id    uint64
	guard *symbols
	count int
}

func (r *rules) first() *symbols { return r.guard.next }
func (r *rules) last() *symbols  { return r.guard.prev }

func (g *Grammar) newRules() *rules {
	r := &rules{id: g.nextID()}
	r.guard = g.newGuard(r)
	return r
}

func (g *Grammar) newSymbolFromValue(sym uint64) *symbols {
	return &symbols{
		g:     g,
		value: sym,
	}
}

func (g *Grammar) newSymbolFromRule(r *rules) *symbols {
	r.count++
	return &symbols{
		g:     g,
		value: r.id,
		rule:  r,
	}
}

func (g *Grammar) newGuard(r *rules) *symbols {
	s := &symbols{g: g, value: r.id, rule: r}
	s.next, s.prev = s, s
	return s
}

func (g *Grammar) newSymbol(s *symbols) *symbols {
	if s.isNonTerminal() {
		return g.newSymbolFromRule(s.rule)
	}
	return g.newSymbolFromValue(s.value)
}

type symbols struct {
	g          *Grammar
	next, prev *symbols
	value      uint64
	rule       *rules
}

func (s *symbols) isGuard() (b bool)   { return s.isNonTerminal() && s.rule.first().prev == s }
func (s *symbols) isNonTerminal() bool { return s.rule != nil }

func (s *symbols) delete() {
	s.prev.join(s.next)
	if s.isGuard() {
		return
	}
	s.deleteDigram()
	if s.isNonTerminal() {
		s.rule.count--
	}
}

func (s *symbols) join(right *symbols) {
	if s.next != nil {
		s.deleteDigram()

		if right.prev != nil && right.next != nil &&
			right.value == right.prev.value &&
			right.value == right.next.value {
			s.g.table.insert(right)
		}

		if s.prev != nil && s.next != nil &&
			s.value == s.next.value &&
			s.value == s.prev.value {
			s.g.table.insert(s.prev)
		}
	}
	s.next = right
	right.prev = s
}

func (s *symbols) insertAfter(y *symbols) {
	y.join(s.next)
	s.join(y)
}

func (s *symbols) deleteDigram() {
	if s.isGuard() || s.next.isGuard() {
		return
	}
	s.g.table.delete(s)
}

func (s *symbols) check() bool {
	if s.isGuard() || s.next.isGuard() {
		return false
	}

	x, ok := s.g.table.lookup(s)
	if !ok {
		s.g.table.insert(s)
		return false
	}

	if x.next != s {
		s.match(x)
	}

	return true
}

func (s *symbols) expand() {
	left := s.prev
	right := s.next
	f := s.rule.first()
	l := s.rule.last()

	s.g.table.delete(s)

	left.join(f)
	l.join(right)

	s.g.table.insert(l)
}

func (s *symbols) substitute(r *rules) {
	q := s.prev

	q.next.delete()
	q.next.delete()

	q.insertAfter(s.g.newSymbolFromRule(r))

	if !q.check() {
		q.next.check()
	}
}

func (s *symbols) match(m *symbols) {
	var r *rules

	if m.prev.isGuard() && m.next.next.isGuard() {
		r = m.prev.rule
		s.substitute(r)
	} else {
		r = s.g.newRules()

		r.last().insertAfter(s.g.newSymbol(s))
		r.last().insertAfter(s.g.newSymbol(s.next))

		m.substitute(r)
		s.substitute(r)

		s.g.table.insert(r.first())
	}

	if r.first().isNonTerminal() && r.first().rule.count == 1 {
		r.first().expand()
	}
}

type digram struct{ one, two uint64 }

type digrams map[digram]*symbols

func (t digrams) lookup(s *symbols) (*symbols, bool) {
	d := digram{s.value, s.next.value}
	m, ok := t[d]
	return m, ok
}

func (t digrams) insert(s *symbols) {
	d := digram{s.value, s.next.value}
	t[d] = s
}

func (t digrams) delete(s *symbols) {
	d := digram{s.value, s.next.value}
	if m, ok := t[d]; ok && s == m {
		delete(t, d)
	}
}

type prettyPrinter struct {
	rules []*rules
	index map[*rules]int
}

func (pr *prettyPrinter) print(w io.Writer, r *rules) error {
	for p := r.first(); !p.isGuard(); p = p.next {
		if p.isNonTerminal() {
			if err := pr.printNonTerminal(w, p.rule); err != nil {
				return err
			}
		} else {
			if err := pr.printTerminal(w, p.value); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func (pr *prettyPrinter) printNonTerminal(w io.Writer, r *rules) error {
	var i int

	if idx, ok := pr.index[r]; ok {
		i = idx
	} else {
		i = len(pr.rules)
		pr.index[r] = i
		pr.rules = append(pr.rules, r)
	}

	_, err := fmt.Fprint(w, i, " ")
	return err
}

func (pr *prettyPrinter) printTerminal(w io.Writer, sym uint64) error {
	var out string

	switch sym {
	case ' ':
		out = "_"
	case '\n':
		out = "\\n"
	case '\t':
		out = "\\t"
	case '\\', '(', ')', '_', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		out = "\\" + string(sym)
	default:
		if _, err := w.Write([]byte{byte(sym)}); err != nil {
			return err
		}
		// leave out empty
	}

	_, err := fmt.Fprint(w, out, " ")
	return err
}

// ErrNoParsedGrammar is returned if no grammar has been parsed
var ErrNoParsedGrammar = errors.New("sequitur: no parsed grammar")

func rawPrint(w io.Writer, r *rules) error {
	for p := r.first(); !p.isGuard(); p = p.next {
		if p.isNonTerminal() {
			if err := rawPrint(w, p.rule); err != nil {
				return err
			}
		} else {
			if _, err := w.Write([]byte{byte(p.value)}); err != nil {
				return err
			}
		}
	}
	return nil
}

// Print reconstructs the input to w
func (g *Grammar) Print(w io.Writer) error {
	if g.base == nil {
		return ErrNoParsedGrammar
	}
	return rawPrint(w, g.base)
}

// PrettyPrint outputs the grammar to w
func (g *Grammar) PrettyPrint(w io.Writer) error {

	if g.base == nil {
		return ErrNoParsedGrammar
	}

	pr := prettyPrinter{
		index: make(map[*rules]int),
		rules: []*rules{g.base},
	}

	for i := 0; i < len(pr.rules); i++ {
		if _, err := fmt.Fprint(w, i, " -> "); err != nil {
			return err
		}

		if err := pr.print(w, pr.rules[i]); err != nil {
			return err
		}
	}

	return nil
}

// ErrAlreadyParsed is returned if the grammar instance has already parsed a grammar
var ErrAlreadyParsed = errors.New("sequitor: grammar already parsed")

// Parse parses a byte string
func (g *Grammar) Parse(str []byte) error {
	if g.base != nil {
		return ErrAlreadyParsed
	}

	g.ruleID = 256
	g.table = make(digrams)
	g.base = g.newRules()

	g.base.last().insertAfter(g.newSymbolFromValue(uint64(str[0])))

	for _, c := range str[1:] {
		g.base.last().insertAfter(g.newSymbolFromValue(uint64(c)))
		g.base.last().prev.check()
	}

	return nil
}
