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
	"unsafe"
)

// Grammar is a constructed grammar.  The zero value is safe to call Parse on.
type Grammar struct {
	table digrams
	base  *rules
}

type rules struct {
	guard *symbols
	count int
}

func (r *rules) reuse() { r.count++ }
func (r *rules) deuse() { r.count-- }

func (r *rules) first() *symbols { return r.guard.next }
func (r *rules) last() *symbols  { return r.guard.prev }

func (r *rules) freq() int { return r.count }

func (g *Grammar) newRules() *rules {
	var r rules

	r.guard = g.newSymbolFromRule(&r)
	r.guard.pointToSelf()
	// r.count is incremented in newSymbolFromRule, but we need to reset it to 0
	r.count = 0

	return &r
}

func (r *rules) delete() {
	r.guard.delete()
}

type symbols struct {
	g          *Grammar
	next, prev *symbols
	value      uintptr
	rule       *rules
}

func (g *Grammar) newSymbolFromValue(sym uintptr) *symbols {
	return &symbols{
		g:     g,
		value: sym,
	}
}

func (g *Grammar) newSymbolFromRule(r *rules) *symbols {
	r.reuse()
	return &symbols{
		g:     g,
		value: uintptr(unsafe.Pointer(r)),
		rule:  r,
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

func (s *symbols) delete() {
	s.prev.join(s.next)
	if !s.isGuard() {
		s.deleteDigram()
		if s.isNonTerminal() {
			s.rule.deuse()
		}
	}
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

func (s *symbols) isGuard() (b bool) {
	return s.isNonTerminal() && s.rule.first().prev == s
}

func (s *symbols) isNonTerminal() bool {
	return s.rule != nil
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

func (s *symbols) pointToSelf() { s.join(s) }

func (s *symbols) expand() {
	left := s.prev
	right := s.next
	f := s.rule.first()
	l := s.rule.last()

	s.rule.delete()
	s.g.table.delete(s)

	s.rule = nil
	s.delete()

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

		if s.isNonTerminal() {
			r.last().insertAfter(s.g.newSymbolFromRule(s.rule))
		} else {
			r.last().insertAfter(s.g.newSymbolFromValue(s.value))
		}

		if s.next.isNonTerminal() {
			r.last().insertAfter(s.g.newSymbolFromRule(s.next.rule))
		} else {
			r.last().insertAfter(s.g.newSymbolFromValue(s.next.value))
		}

		m.substitute(r)
		s.substitute(r)

		s.g.table.insert(r.first())
	}

	if r.first().isNonTerminal() && r.first().rule.freq() == 1 {
		r.first().expand()
	}
}

type digram struct {
	one, two uintptr
}

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

type Printer struct {
	rules []*rules
	index map[*rules]int
}

func (pr *Printer) print(w io.Writer, r *rules) error {
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
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

func (pr *Printer) printNonTerminal(w io.Writer, r *rules) error {
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

func (pr *Printer) printTerminal(w io.Writer, sym uintptr) error {
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

var ErrNoParsedGrammar = errors.New("sequitor: no parsed grammar")

// Print outputs the grammar to w
func (g *Grammar) Print(w io.Writer) error {

	if g.base == nil {
		return ErrNoParsedGrammar
	}

	pr := Printer{
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

	g.table = make(digrams)
	g.base = g.newRules()

	g.base.last().insertAfter(g.newSymbolFromValue(uintptr(str[0])))

	for _, c := range str[1:] {
		g.base.last().insertAfter(g.newSymbolFromValue(uintptr(c)))
		g.base.last().prev.check()
	}

	return nil
}
