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
	"unicode"
	"unicode/utf8"
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
	s.deleteDigram()
	if s.isNonTerminal() {
		s.rule.count--
	}
}

func (s *symbols) isTriple() bool {
	return s.prev != nil && s.next != nil &&
		s.value == s.prev.value &&
		s.value == s.next.value
}

func (s *symbols) join(right *symbols) {
	if s.next != nil {
		s.deleteDigram()

		if right.isTriple() {
			s.g.table.insert(right)
		}

		if s.isTriple() {
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

	_, err := fmt.Fprint(w, " ", i)
	return err
}

func (pr *prettyPrinter) printTerminal(w io.Writer, sym uint64) error {
	out := make([]byte, 1, 1+utf8.UTFMax)
	out[0] = ' '
	rb := runeOrByte(sym)
	switch r := rb.rune(); r {
	case ' ':
		out = append(out, '_')
	case '\n':
		out = append(out, []byte("\\n")...)
	case '\t':
		out = append(out, []byte("\\t")...)
	case '\\', '(', ')', '_', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		out = append(out, '\\', byte(r))
	default:
		out = rb.appendEscaped(out)
	}
	_, err := w.Write(out)
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
			if _, err := w.Write(runeOrByte(p.value).appendBytes(nil)); err != nil {
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
		if _, err := fmt.Fprint(w, i, " ->"); err != nil {
			return err
		}

		if err := pr.print(w, pr.rules[i]); err != nil {
			return err
		}
	}

	return nil
}

// ErrEmptyInput is returned if the input string is empty
var ErrEmptyInput = errors.New("sequitur: empty input")

// Parse parses the given bytes.
func Parse(str []byte) (*Grammar, error) {
	if len(str) == 0 {
		return nil, ErrEmptyInput
	}
	g := &Grammar{
		ruleID: uint64(utf8.MaxRune) + 1, // larger than the largest rune
		table:  make(digrams),
	}
	g.base = g.newRules()
	for off := 0; off < len(str); {
		var rb runeOrByte
		r, sz := utf8.DecodeRune(str[off:])
		if sz == 1 && r == utf8.RuneError {
			rb = newByte(str[off])
		} else {
			rb = newRune(r)
		}
		g.base.last().insertAfter(g.newSymbolFromValue(uint64(rb)))
		if off > 0 {
			g.base.last().prev.check()
		}
		off += sz
	}
	return g, nil
}

// runeOrByte holds a rune or a byte so that we can distinguish between
// bytes that don't represent valid UTF-8 and all other runes. Values
// not representable as UTF-8 are in the range 128-255. All other
// runes are represented as 256 onwards (subtract 256 to get the
// actual rune value). Note that the range 0-127 is unused.
type runeOrByte rune

func newRune(r rune) runeOrByte {
	return runeOrByte(r + 256)
}

// newByte returns a representation of the given byte b.
func newByte(b byte) runeOrByte {
	if b < utf8.RuneSelf {
		return runeOrByte(b) + 256
	}
	return runeOrByte(b)
}

// rune returns the rune representation of
// rb, or zero if there is none.
func (rb runeOrByte) rune() rune {
	if rb < 256 {
		return 0
	}
	return rune(rb - 256)
}

// appendEscaped appends the possibly escaped rune or byte
// to b. If it's printable, the printable representation is appended,
// otherwise \x, \u or \U are used as appropriate.
// Note, it doesn't escape \ itself.
func (rb runeOrByte) appendEscaped(b []byte) []byte {
	if rb < 256 {
		return append(b, fmt.Sprintf("\\x%02x", rb)...)
	}
	r := rune(rb - 256)
	switch {
	case unicode.IsPrint(r):
		return append(b, string(r)...)
	case r < utf8.RuneSelf:
		// Could use either representation, but \x is shorter.
		return append(b, fmt.Sprintf("\\x%02x", r)...)
	case r <= 0xffff:
		return append(b, fmt.Sprintf("\\u%04x", r)...)
	default:
		return append(b, fmt.Sprintf("\\U%08x", r)...)
	}
}

// appendBytes appends the byte (as a byte) or the rune (as utf-8)
// to b.
func (rb runeOrByte) appendBytes(b []byte) []byte {
	if rb < 256 {
		return append(b, byte(rb))
	}
	return append(b, string(rb-256)...)
}
