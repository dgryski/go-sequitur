package sequitur

import (
	"fmt"
	"io"
	"unsafe"
)

type rules struct {
	guard  *symbols
	count  int
	number int
}

func (r *rules) reuse() { r.count++ }
func (r *rules) deuse() { r.count-- }

func (r *rules) first() *symbols { return r.guard.next() }
func (r *rules) last() *symbols  { return r.guard.prev() }

func (r *rules) freq() int      { return r.count }
func (r *rules) index() int     { return r.number }
func (r *rules) setIndex(i int) { r.number = i }

func newRules() *rules {
	var r rules

	r.guard = newSymbolFromRule(&r)
	r.guard.point_to_self()
	r.count = 0
	r.number = 0

	return &r
}

func (r *rules) delete() {
	r.guard.delete()
}

type symbols struct {
	n, p *symbols
	s    uintptr
	r    *rules
}

func newSymbolFromValue(sym uintptr) *symbols {
	return &symbols{
		s: sym,
	}
}

func newSymbolFromRule(r *rules) *symbols {
	r.reuse()
	return &symbols{
		s: uintptr(unsafe.Pointer(r)),
		r: r,
	}
}

func (left *symbols) join(right *symbols) {
	if left.n != nil {
		left.delete_digram()

		if right.p != nil && right.n != nil &&
			right.value() == right.p.value() &&
			right.value() == right.n.value() {
			set_digram(right)
		}

		if left.p != nil && left.n != nil &&
			left.value() == left.n.value() &&
			left.value() == left.p.value() {
			set_digram(left.p)
		}
	}
	left.n = right
	right.p = left
}

func (s *symbols) delete() {
	s.p.join(s.n)
	if !s.is_guard() {
		s.delete_digram()
		if s.nt() {
			s.rule().deuse()
		}
	}
}

func (s *symbols) insert_after(y *symbols) {
	y.join(s.n)
	s.join(y)
}

func (s *symbols) delete_digram() {
	if s.is_guard() || s.n.is_guard() {
		return
	}
	delete_from_hash_table(s)
}

func (s *symbols) is_guard() (b bool) {
	return s.nt() && s.rule().first().prev() == s
}

func (s *symbols) nt() bool {
	return s.r != nil
}

func (s *symbols) next() *symbols { return s.n }
func (s *symbols) prev() *symbols { return s.p }

func (s *symbols) raw_value() uintptr {

	if s.r != nil && uintptr(s.s) != uintptr(unsafe.Pointer(s.r)) {
		panic("s.r is non nil but s.s doesn't match")
	}

	return uintptr(s.s)
}
func (s *symbols) value() uintptr {
	return uintptr(s.s)
}

func (s *symbols) rule() *rules { return s.r }

func (s *symbols) check() bool {
	if s.is_guard() || s.n.is_guard() {
		return false
	}

	x, ok := find_digram(s)
	if !ok {
		set_digram(s)
		return false
	}

	if x.next() != s {
		match(s, x)
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
	delete_from_hash_table(s)

	s.r = nil
	s.delete()

	left.join(f)
	l.join(right)

	set_digram(l)
}

func (s *symbols) substitute(r *rules) {
	q := s.p

	q.next().delete()
	q.next().delete()

	q.insert_after(newSymbolFromRule(r))

	if !q.check() {
		q.n.check()
	}
}

func match(ss, m *symbols) {

	var r *rules

	if m.prev().is_guard() && m.next().next().is_guard() {
		r = m.prev().rule()
		ss.substitute(r)
	} else {
		r = newRules()

		if ss.nt() {
			r.last().insert_after(newSymbolFromRule(ss.rule()))
		} else {
			r.last().insert_after(newSymbolFromValue(ss.value()))
		}

		if ss.next().nt() {
			r.last().insert_after(newSymbolFromRule(ss.next().rule()))
		} else {
			r.last().insert_after(newSymbolFromValue(ss.next().value()))
		}

		m.substitute(r)
		ss.substitute(r)

		set_digram(r.first())
	}

	if r.first().nt() && r.first().rule().freq() == 1 {
		r.first().expand()
	}
}

type digram struct {
	one, two uintptr
}

var table = map[digram]*symbols{}

func find_digram(s *symbols) (*symbols, bool) {
	one := s.raw_value()
	two := s.next().raw_value()
	d := digram{one, two}
	m, ok := table[d]
	if ok {
		if m.raw_value() != one || m.next().raw_value() != two {
			panic("hash table mismatch")
		}
	}
	return m, ok
}

func set_digram(s *symbols) {
	one := s.raw_value()
	two := s.next().raw_value()
	d := digram{one, two}
	table[d] = s
}

func delete_from_hash_table(s *symbols) {
	one := s.raw_value()
	two := s.next().raw_value()
	d := digram{one, two}
	if m, ok := table[d]; ok && s == m {
		delete(table, d)
	}
}

type Printer struct {
	R  []*rules
	Ri int
}

func (pr *Printer) print(w io.Writer, r *rules) {
	for p := r.first(); !p.is_guard(); p = p.next() {
		if p.nt() {
			var i int

			if p.rule().index() < len(pr.R) && pr.R[p.rule().index()] == p.rule() {
				i = p.rule().index()
			} else {
				i = pr.Ri
				p.rule().setIndex(pr.Ri)
				pr.R = append(pr.R, p.rule())
				pr.Ri++
			}

			fmt.Fprint(w, i, " ")
		} else {
			if p.value() == ' ' {
				fmt.Fprint(w, "_")
			} else if p.value() == '\n' {
				fmt.Fprint(w, "\\n")
			} else if p.value() == '\t' {
				fmt.Fprint(w, "\\t")
			} else if p.value() == '\\' ||
				p.value() == '(' ||
				p.value() == ')' ||
				p.value() == '_' ||
				isdigit(p.value()) {
				fmt.Fprint(w, string([]byte{'\\', byte(p.value())}))
			} else {
				w.Write([]byte{byte(p.value())})
				//fmt.Fprintf(w, "%s", string(byte(p.value())))
			}
			fmt.Fprint(w, " ")
		}
	}
	fmt.Fprintln(w)
}

func isdigit(c uintptr) bool { return c >= '0' && c <= '9' }

func (pr *Printer) Print(w io.Writer, r *rules) {
	pr.R = []*rules{r}
	pr.Ri = 1

	for i := 0; i < pr.Ri; i++ {
		fmt.Fprint(w, i, " -> ")
		pr.print(w, pr.R[i])
	}

	for _, v := range pr.R {
		if v != nil {
			v.number = 0
		}
	}
}

func ParseAndPrint(w io.Writer, str []byte) {

	// reset global state
	table = make(map[digram]*symbols)

	S := newRules()
	S.last().insert_after(newSymbolFromValue(uintptr(str[0])))

	for _, c := range str[1:] {
		S.last().insert_after(newSymbolFromValue(uintptr(c)))
		S.last().prev().check()
		//	fmt.Fprintf(w, "R=%v\n", R[:Ri])
		//	fmt.Fprintf(w, "table=%v\n", table)
		//	print(w, S)
	}

	var pr Printer
	pr.Print(w, S)
}
