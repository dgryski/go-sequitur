package sequitur

import (
	"errors"
	"fmt"
	"io"
	"unsafe"
)

type Grammar struct {
	table digrams
	base  *rules
}

func NewGrammar() *Grammar {
	return &Grammar{
		table: make(digrams),
	}
}

type rules struct {
	guard *symbols
	count int
}

func (r *rules) reuse() { r.count++ }
func (r *rules) deuse() { r.count-- }

func (r *rules) first() *symbols { return r.guard.next() }
func (r *rules) last() *symbols  { return r.guard.prev() }

func (r *rules) freq() int { return r.count }

func (g *Grammar) newRules() *rules {
	var r rules

	r.guard = g.newSymbolFromRule(&r)
	r.guard.point_to_self()
	// r.count is incremented in newSymbolFromRule, but we need to reset it to 0
	r.count = 0

	return &r
}

func (r *rules) delete() {
	r.guard.delete()
}

type symbols struct {
	g    *Grammar
	n, p *symbols
	s    uintptr
	r    *rules
}

func (g *Grammar) newSymbolFromValue(sym uintptr) *symbols {
	return &symbols{
		g: g,
		s: sym,
	}
}

func (g *Grammar) newSymbolFromRule(r *rules) *symbols {
	r.reuse()
	return &symbols{
		g: g,
		s: uintptr(unsafe.Pointer(r)),
		r: r,
	}
}

func (s *symbols) join(right *symbols) {
	if s.n != nil {
		s.deleteDigram()

		if right.p != nil && right.n != nil &&
			right.value() == right.p.value() &&
			right.value() == right.n.value() {
			s.g.table.insert(right)
		}

		if s.p != nil && s.n != nil &&
			s.value() == s.n.value() &&
			s.value() == s.p.value() {
			s.g.table.insert(s.p)
		}
	}
	s.n = right
	right.p = s
}

func (s *symbols) delete() {
	s.p.join(s.n)
	if !s.isGuard() {
		s.deleteDigram()
		if s.isNonTerminal() {
			s.rule().deuse()
		}
	}
}

func (s *symbols) insertAfter(y *symbols) {
	y.join(s.n)
	s.join(y)
}

func (s *symbols) deleteDigram() {
	if s.isGuard() || s.n.isGuard() {
		return
	}
	s.g.table.delete(s)
}

func (s *symbols) isGuard() (b bool) {
	return s.isNonTerminal() && s.rule().first().prev() == s
}

func (s *symbols) isNonTerminal() bool {
	return s.r != nil
}

func (s *symbols) next() *symbols { return s.n }
func (s *symbols) prev() *symbols { return s.p }

func (s *symbols) value() uintptr { return s.s }

func (s *symbols) rule() *rules { return s.r }

func (s *symbols) check() bool {
	if s.isGuard() || s.n.isGuard() {
		return false
	}

	x, ok := s.g.table.lookup(s)
	if !ok {
		s.g.table.insert(s)
		return false
	}

	if x.next() != s {
		s.match(x)
	}

	return true
}

func (s *symbols) point_to_self() { s.join(s) }

func (s *symbols) expand() {
	left := s.prev()
	right := s.next()
	f := s.rule().first()
	l := s.rule().last()

	s.rule().delete()
	s.g.table.delete(s)

	s.r = nil
	s.delete()

	left.join(f)
	l.join(right)

	s.g.table.insert(l)
}

func (s *symbols) substitute(r *rules) {
	q := s.p

	q.next().delete()
	q.next().delete()

	q.insertAfter(s.g.newSymbolFromRule(r))

	if !q.check() {
		q.n.check()
	}
}

func (s *symbols) match(m *symbols) {

	var r *rules

	if m.prev().isGuard() && m.next().next().isGuard() {
		r = m.prev().rule()
		s.substitute(r)
	} else {
		r = s.g.newRules()

		if s.isNonTerminal() {
			r.last().insertAfter(s.g.newSymbolFromRule(s.rule()))
		} else {
			r.last().insertAfter(s.g.newSymbolFromValue(s.value()))
		}

		if s.next().isNonTerminal() {
			r.last().insertAfter(s.g.newSymbolFromRule(s.next().rule()))
		} else {
			r.last().insertAfter(s.g.newSymbolFromValue(s.next().value()))
		}

		m.substitute(r)
		s.substitute(r)

		s.g.table.insert(r.first())
	}

	if r.first().isNonTerminal() && r.first().rule().freq() == 1 {
		r.first().expand()
	}
}

type digram struct {
	one, two uintptr
}

type digrams map[digram]*symbols

func (t digrams) lookup(s *symbols) (*symbols, bool) {
	one := s.value()
	two := s.next().value()
	d := digram{one, two}
	m, ok := t[d]
	return m, ok
}

func (t digrams) insert(s *symbols) {
	one := s.value()
	two := s.next().value()
	d := digram{one, two}
	t[d] = s
}

func (t digrams) delete(s *symbols) {
	one := s.value()
	two := s.next().value()
	d := digram{one, two}
	if m, ok := t[d]; ok && s == m {
		delete(t, d)
	}
}

type Printer struct {
	rules []*rules
	index map[*rules]int
}

func (pr *Printer) print(w io.Writer, r *rules) {
	for p := r.first(); !p.isGuard(); p = p.next() {
		if p.isNonTerminal() {
			pr.printNonTerminal(w, p.rule())
		} else {
			pr.printTerminal(w, p.value())
		}
	}
	fmt.Fprintln(w)
}

func (pr *Printer) printNonTerminal(w io.Writer, r *rules) {
	var i int

	if idx, ok := pr.index[r]; ok {
		i = idx
	} else {
		i = len(pr.rules)
		pr.index[r] = i
		pr.rules = append(pr.rules, r)
	}

	fmt.Fprint(w, i, " ")
}

func (pr *Printer) printTerminal(w io.Writer, sym uintptr) {
	if sym == ' ' {
		fmt.Fprint(w, "_")
	} else if sym == '\n' {
		fmt.Fprint(w, "\\n")
	} else if sym == '\t' {
		fmt.Fprint(w, "\\t")
	} else if sym == '\\' ||
		sym == '(' ||
		sym == ')' ||
		sym == '_' ||
		isdigit(sym) {
		fmt.Fprint(w, string([]byte{'\\', byte(sym)}))
	} else {
		w.Write([]byte{byte(sym)})
	}
	fmt.Fprint(w, " ")
}

func isdigit(c uintptr) bool { return c >= '0' && c <= '9' }

func (g *Grammar) Print(w io.Writer) {
	pr := Printer{
		index: make(map[*rules]int),
		rules: []*rules{g.base},
	}

	for i := 0; i < len(pr.rules); i++ {
		fmt.Fprint(w, i, " -> ")
		pr.print(w, pr.rules[i])
	}
}

var ErrAlreadyParsed = errors.New("sequitor: grammar already parsed")

func (g *Grammar) Parse(str []byte) error {
	if g.base != nil {
		return ErrAlreadyParsed
	}

	g.table = make(digrams)
	g.base = g.newRules()

	g.base.last().insertAfter(g.newSymbolFromValue(uintptr(str[0])))

	for _, c := range str[1:] {
		g.base.last().insertAfter(g.newSymbolFromValue(uintptr(c)))
		g.base.last().prev().check()
	}

	return nil
}
