package sequitur

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"unicode/utf8"
)

// Symbol is the named type for a grammar symbol.
// A nil value for *Symbol means it is representing an empty grammar.
type Symbol symbols

// String for Symbol.
func (s *Symbol) String() string {
	return s.ID().String()
}

// SymbolID is a rune encoding the symbol.
// There are 2,146,369,280 possible +ve values of rune above maxRuneOrByte which are used for rule IDs (composite symbols).
type SymbolID rune

const EmptySymbolID = -1 // Shows that a SymbolID represents an empty grammar.

const EmptySymbolIDstring = "\\empty" // The visual representation of an empty grammar.

// String for SymbolID.
func (sid SymbolID) String() string {
	if sid == EmptySymbolID {
		return EmptySymbolIDstring
	}
	if sid.IsRule() {
		return fmt.Sprint(uint32(sid))
	}
	return string(runeOrByte(sid).appendEscaped(make([]byte, 0, utf8.UTFMax)))
}

// IsRule says if the id originated as a rule rather than a runeOrByte.
func (sid SymbolID) IsRule() bool {
	return sid > SymbolID(maxRuneOrByte)
}

// ID of a Symbol.
func (s *Symbol) ID() SymbolID {
	if s == nil {
		return EmptySymbolID
	}
	if s.rule != nil {
		return SymbolID(s.rule.id)
	}
	return SymbolID(s.value)
}

// Bytes of a Symbol and all its sub-symbols.
func (s *Symbol) Bytes() []byte {
	if s == nil {
		return nil
	}
	if s.rule != nil {
		var b bytes.Buffer
		_ = rawPrint(&b, s.rule) // ignore error
		return b.Bytes()
	}
	return runeOrByte(s.value).appendBytes(make([]byte, 0, utf8.UTFMax))
}

// Used gives the number of times this symbol has been reused.
func (s *Symbol) Used() int {
	if s == nil {
		return 0
	}
	if s.rule != nil {
		return s.rule.count
	}
	return 1
}

// SubSymbols returns a slice of the symbols comprising this one.
// It returns an empty slice if the symbol is not a rule.
func (s *Symbol) SubSymbols() []*Symbol {
	if s == nil {
		return nil
	}
	ret := []*Symbol{}
	if s.rule != nil {
		for p := s.rule.first(); !p.isGuard(); p = p.next {
			ret = append(ret, (*Symbol)(p))
		}
	}
	return ret
}

// Symbol provides the top-level symbol for a Grammar.
func (g *Grammar) Symbol() *Symbol {
	ret := (*Symbol)(&symbols{
		g:    g,
		rule: g.base,
	})
	if len(ret.SubSymbols()) == 0 {
		return nil // a nil *Symbol represents an empty grammar.
	}
	return ret
}

// Compact provides a more compact representation of the grammar, making it suitable for serialisation.
// Only SymbolIDs with IsRule()==true have entries in the map.
type Compact struct {
	RootID SymbolID
	Map    map[SymbolID]CompactEntry
}

// String form of a Compact grammar, returns .PrettyPrint() output or "\empty".
func (comp *Compact) String() string {
	if comp.RootID == EmptySymbolID {
		return comp.RootID.String()
	}
	var b bytes.Buffer
	if err := comp.PrettyPrint(&b); err != nil {
		return err.Error()
	}
	return b.String()
}

// CompactEntry gives the minimal information about a symbol which is comprised of others.
type CompactEntry struct {
	Used int
	IDs  SymbolIDslice
}

// SymbolIDslice is a slice of SymbolIDs.
type SymbolIDslice []SymbolID

// Compact returns the Compact representation of a Grammar.
func (g *Grammar) Compact() *Compact {
	gs := g.Symbol()
	id := gs.ID()
	fm := &Compact{
		RootID: id,
		Map:    make(map[SymbolID]CompactEntry),
	}
	if id != EmptySymbolID {
		fm.addSymbol(gs)
	}
	return fm
}

func (comp *Compact) addSymbol(s *Symbol) {
	if comp == nil {
		return
	}
	id := s.ID()
	if id.IsRule() {
		_, exists := comp.Map[id]
		if !exists {
			subSyms := s.SubSymbols()
			entry := CompactEntry{
				Used: s.Used(),
				IDs:  make([]SymbolID, len(subSyms)),
			}
			for k, v := range subSyms {
				entry.IDs[k] = v.ID()
				comp.addSymbol(v)
			}
			comp.Map[id] = entry
		}
	}
}

// Bytes of a SymbolID, including all of the symbols that it contains.
func (sid SymbolID) Bytes(comp *Compact) []byte {
	if sid == EmptySymbolID || comp == nil {
		return nil
	}
	if sid.IsRule() {
		entry := comp.Map[sid]
		result := make([]byte, 0, len(entry.IDs)*utf8.UTFMax)
		for _, eid := range entry.IDs {
			result = append(result, eid.Bytes(comp)...)
		}
		return result
	}
	return runeOrByte(sid).appendBytes(make([]byte, 0, utf8.UTFMax))
}

// Bytes of a SymbolIDslice, including all of the symbols that it contains.
func (sids SymbolIDslice) Bytes(comp *Compact) []byte {
	if len(sids) == 0 || comp == nil {
		return nil
	}
	result := make([]byte, 0, len(sids)*utf8.UTFMax)
	for _, id := range sids {
		result = append(result, id.Bytes(comp)...)
	}
	return result
}

func (comp *Compact) prettyPrint(id SymbolID, seenMap map[SymbolID]string) {
	entry := comp.Map[id]
	for _, ss := range entry.IDs {
		if ss.IsRule() {
			_, seen := seenMap[ss]
			if !seen {
				comp.prettyPrint(ss, seenMap)
			}
		}
	}
	seenMap[id] = fmt.Sprintln(int32(id), "->", entry)
}

// PrettyPrint a Compact grammar, using actual IDs.
func (comp *Compact) PrettyPrint(w io.Writer) error {
	if comp == nil {
		return nil
	}
	if comp.RootID == EmptySymbolID {
		return nil
	}
	output := make(map[SymbolID]string)
	comp.prettyPrint(comp.RootID, output)
	idList := make(SymbolIDslice, 0, len(output))
	for id := range output {
		idList = append(idList, id)
	}
	sort.Slice(idList, func(i, j int) bool { return idList[i] < idList[j] })
	for _, id := range idList {
		if _, err := w.Write([]byte(output[id])); err != nil {
			return err
		}
	}
	return nil
}

// Bytes of a Compact grammar SymbolID, including all of the symbols that it contains.
func (comp *Compact) Bytes(sid SymbolID) []byte {
	if sid == EmptySymbolID || comp == nil {
		return nil
	}
	if uint64(sid) <= maxRuneOrByte {
		return runeOrByte(sid).appendBytes(make([]byte, 0, utf8.UTFMax))
	}
	return comp.Map[sid].IDs.Bytes(comp)
}

// CompactIndexes indexes the Compact datastructure.
type CompactIndexed struct {
	CompactBasis        *Compact
	MinSymByteLen       int
	TrimSpace           bool
	OriginalInputLength int
	TotalCoverage       float64
	StringToID          map[string]SymbolID
	IDinfo              map[SymbolID]CompactIndexedInfo
}

// CompactIndexedInfo stores derrived information about a Symbol.
type CompactIndexedInfo struct {
	Coverage float64 // the proportion of the original input represented by this symbol
}

// Index the Compact grammar to enable further analysis, optionally filtering the []byte representations of the symbols.
func (comp *Compact) Index(filterKeep func([]byte) bool) *CompactIndexed {
	if comp == nil {
		return nil
	}
	ret := &CompactIndexed{
		CompactBasis: comp,
		StringToID:   make(map[string]SymbolID),
		IDinfo:       make(map[SymbolID]CompactIndexedInfo),
	}
	if filterKeep == nil {
		filterKeep = func([]byte) bool { return true }
	}
	for k, v := range comp.Map {
		b := v.IDs.Bytes(comp)
		if k == comp.RootID {
			ret.OriginalInputLength = len(b)
		}
		if filterKeep(b) {
			ret.StringToID[string(b)] = k
			ret.IDinfo[k] = CompactIndexedInfo{
				Coverage: float64(len(b)),
			}
		}
	}
	for k, v := range ret.IDinfo {
		v.Coverage /= float64(ret.OriginalInputLength)
		ret.IDinfo[k] = v
		ret.TotalCoverage += v.Coverage
	}
	return ret
}

type Importance struct {
	ID    SymbolID
	Score float64
}

// Importance ranks the most important IDs according to the given scoring function, or the coverage if the function is nil.
func (ci *CompactIndexed) Importance(scoreFn func(SymbolID) float64) []Importance {
	if ci == nil {
		return nil
	}
	imp := make([]Importance, 0, len(ci.CompactBasis.Map))
	for k, v := range ci.IDinfo {
		score := v.Coverage
		if scoreFn != nil {
			score = scoreFn(k)
		}
		imp = append(imp, Importance{
			ID:    k,
			Score: score,
		})
	}
	sort.Slice(imp, func(i, j int) bool {
		if imp[i].Score == imp[j].Score {
			return imp[i].ID > imp[j].ID // arbritrary but stable order
		}
		return imp[i].Score > imp[j].Score
	})
	return imp
}

// Similarity between two CompactIndexed grammars. Result: 1 (or nearby) equality, 0 inequality.
func (ci *CompactIndexed) Similarity(ci2 *CompactIndexed) float64 {
	if ci == nil || ci2 == nil {
		return 0
	}
	cumCoverage := 0.0
	if len(ci2.StringToID) < len(ci.StringToID) {
		ci2, ci = ci, ci2 // swap to iterate over ci2 if it is shorter
	}
	for str, sid := range ci.StringToID {
		sid2, found := ci2.StringToID[str]
		if found {
			cumCoverage += ci.IDinfo[sid].Coverage + ci2.IDinfo[sid2].Coverage
		}
	}
	divisor := (ci.TotalCoverage + ci2.TotalCoverage)
	if divisor == 0 {
		return 1 // two empty grammars are equal
	}
	return cumCoverage / divisor
}
