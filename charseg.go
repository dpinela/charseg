// Package charseg implements Unicode grapheme cluster segmentation.
//
// The segmentation rules are as defined by Unicode Standard Annex 29:
// https://unicode.org/reports/tr29/
package charseg

import (
	"sort"
	"unicode/utf8"
)

const (
	catEOF     category = -1
	catUnknown          = -2
)

type segmenter struct {
	lastCategory category

	inEmojiSequence, notAtStart bool
	numRISymbols                int
}

// FirstGraphemeCluster returns the first grapheme cluster in a string.
func FirstGraphemeCluster(text string) string {
	var seg segmenter
	return text[:seg.NextBoundary(text, true)]
}

func (s *segmenter) NextBoundary(text string, atEOF bool) int {
	i := 0
	for i < len(text) {
		r, n := utf8.DecodeRuneInString(text[i:])
		c := classify(r)
		if has, sure := s.hasBoundaryBefore(c); sure {
			if has {
				return i
			}
		}
		s.updateState(c)
		i += n
	}
	if has, sure := s.hasBoundaryBefore(eofCategory(atEOF)); sure && has {
		return len(text)
	}
	return -1
}

func eofCategory(atEOF bool) category {
	if atEOF {
		return catEOF
	}
	return catUnknown
}

func (s *segmenter) updateState(y category) {
	s.lastCategory = y
	s.notAtStart = true
	if y == catRegional_Indicator {
		s.numRISymbols++
		return
	}
	s.numRISymbols = 0
	if y == catE_Base || y == catE_Base_GAZ {
		s.inEmojiSequence = true
		return
	}
	if s.inEmojiSequence && y != catExtend {
		s.inEmojiSequence = false
	}
}

func (s *segmenter) hasBoundaryBefore(y category) (found, sure bool) {
	// Ignore the boundary at start of text (GB1); we never return that one.
	if !s.notAtStart {
		return false, true
	}
	// Rule GB2
	if y == catEOF {
		return true, true
	}
	x := s.lastCategory
	// Rule GB3
	if x == catCR {
		switch y {
		case catUnknown:
			return false, false
		case catLF:
			return false, true
		}
		return true, true
	}
	// Rule GB4-5 (no need to check for EOF if x is a control character,
	// because then we don't care what comes after)
	if isControlOrLF(x) || isControlOrLF(y) {
		return true, true
	}
	// Rule GB6-8 (Hangul)
	switch x {
	case catL:
		switch y {
		case catUnknown:
			return false, false
		case catL, catV, catLV, catLVT:
			return false, true
		}
	case catLV, catV:
		switch y {
		case catUnknown:
			return false, false
		case catV, catT:
			return false, true
		}
	case catLVT, catT:
		switch y {
		case catUnknown:
			return false, false
		case catT:
			return false, true
		}
	}
	// Rule GB9, 9a and 9b
	switch y {
	case catUnknown:
		return false, false
	case catExtend, catZWJ, catSpacingMark:
		return false, true
	}
	if x == catPrepend {
		return false, true
	}
	// Rule GB10 (Emoji modifier sequences)
	if s.inEmojiSequence {
		switch y {
		case catUnknown:
			return false, false
		case catE_Modifier:
			return false, true
		}
	}
	// Rule GB11 (Emoji ZWJ sequences)
	if x == catZWJ {
		switch y {
		case catUnknown:
			return false, false
		case catGlue_After_Zwj, catE_Base_GAZ:
			return false, true
		}
	}
	// Rule GB12-13 (Emoji flag sequences)
	if (s.numRISymbols & 1) == 1 {
		switch y {
		case catUnknown:
			return false, false
		case catRegional_Indicator:
			return false, true
		}
	}
	z := y != catUnknown
	return z, z
}

func classify(x rune) category {
	i := sort.Search(len(ranges), func(j int) bool { return x < ranges[j].End })
	if i == len(ranges) || x < ranges[i].Begin {
		return catNone
	}
	return categories[i]
}

func isControlOrLF(x category) bool { return x == catControl || x == catCR || x == catLF }
