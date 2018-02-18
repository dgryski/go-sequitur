package sequitur

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"unicode/utf8"
)

// Symbol is the named type for a grammar symbol.
type Symbol symbols

// String for Symbol.
func (s *Symbol) String() string {
	return s.ID().String()
}

// SymbolID is a rune encoding the symbol.
// There are 2,146,369,280 possible +ve values of rune above maxRuneOrByte which are used for rule IDs (composite symbols).
type SymbolID rune

// String for SymbolID.
func (sid SymbolID) String() string {
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
	if s.rule != nil {
		return SymbolID(s.rule.id)
	}
	return SymbolID(s.value)
}

// Bytes of a Symbol and all its sub-symbols.
func (s *Symbol) Bytes() []byte {
	if s.rule != nil {
		var b bytes.Buffer
		_ = rawPrint(&b, s.rule) // ignore error
		return b.Bytes()
	}
	return runeOrByte(s.value).appendBytes(make([]byte, 0, utf8.UTFMax))
}

// Used gives the number of times this symbol has been reused.
func (s *Symbol) Used() int {
	if s.rule != nil {
		return s.rule.count
	}
	return 1
}

// SubSymbols returns a slice of the symbols comprising this one.
// It returns an empty slice if the symbol is not a rule.
func (s *Symbol) SubSymbols() []*Symbol {
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
	return (*Symbol)(&symbols{
		g:    g,
		rule: g.base,
	})
}

// Compact provides a more compact representation of the grammar, making it suitable for serialisation.
// Only SymbolIDs with IsRule()==true have entries in the map.
type Compact struct {
	RootID SymbolID
	Map    map[SymbolID]CompactEntry
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
	fm := &Compact{
		RootID: gs.ID(),
		Map:    make(map[SymbolID]CompactEntry),
	}
	fm.addSymbol(gs)
	return fm
}

func (comp *Compact) addSymbol(s *Symbol) {
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
	StringToID          map[string]SymbolID
	IDinfo              map[SymbolID]CompactIndexedInfo
}

// CompactIndexedInfo stores derrived information about a Symbol.
type CompactIndexedInfo struct {
	Coverage float64 // the proportion of the original input represented by this symbol
}

// Index the Compact grammar to enable further analysis, optionally filtering the []byte representations of the symbols.
func (comp *Compact) Index(filter func([]byte) []byte) *CompactIndexed {
	ret := &CompactIndexed{
		CompactBasis: comp,
		StringToID:   make(map[string]SymbolID),
		IDinfo:       make(map[SymbolID]CompactIndexedInfo),
	}
	for k, v := range comp.Map {
		bOrig := v.IDs.Bytes(comp)
		if k == comp.RootID {
			ret.OriginalInputLength = len(bOrig)
		}
		b := bOrig
		if filter != nil {
			b = filter(b)
		}
		if len(b) > 0 {
			ret.StringToID[string(b)] = k
			ret.IDinfo[k] = CompactIndexedInfo{
				Coverage: float64(len(bOrig)),
			}
		}
	}
	for k, v := range ret.IDinfo {
		v.Coverage /= float64(ret.OriginalInputLength)
		ret.IDinfo[k] = v
	}
	return ret
}

type Importance struct {
	ID    SymbolID
	Score float64
}

// Importance ranks the most important IDs according to the given scoring function, or the coverage if the function is nil.
func (ci *CompactIndexed) Importance(scoreFn func(SymbolID) float64) []Importance {
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
