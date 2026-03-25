package tui

import "strings"

// FuzzyList implements a reusable fuzzy-search list with cursor navigation.
// Used by CompareView's picker and JumperView.
type FuzzyList struct {
	Items   []PickerItem
	Matches []pickerMatch
	Cursor  int
	Offset  int
	MaxShow int
}

// NewFuzzyList creates a FuzzyList with the given items and visible row limit.
func NewFuzzyList(items []PickerItem, maxShow int) FuzzyList {
	fl := FuzzyList{Items: items, MaxShow: maxShow}
	fl.Filter("")
	return fl
}

// Filter updates the match list based on the query string.
// Supports type-prefixed queries like "mf:query" where the prefix is
// fuzzy-matched against NodeType (e.g. "mf" matches "Microflow").
func (fl *FuzzyList) Filter(query string) {
	query = strings.TrimSpace(query)
	fl.Matches = nil

	// Parse optional type prefix: "prefix:nameQuery"
	typeFilter, nameQuery := "", query
	if idx := strings.IndexByte(query, ':'); idx > 0 && idx < len(query)-1 {
		typeFilter = query[:idx]
		nameQuery = query[idx+1:]
	} else if idx > 0 && idx == len(query)-1 {
		// Trailing colon with no name query — filter by type only
		typeFilter = query[:idx]
		nameQuery = ""
	}

	for _, it := range fl.Items {
		if query == "" {
			fl.Matches = append(fl.Matches, pickerMatch{item: it})
			continue
		}
		// Type filter: fuzzy match prefix against NodeType
		if typeFilter != "" {
			if ok, _ := fuzzyScore(it.NodeType, typeFilter); !ok {
				continue
			}
		}
		// Name query: fuzzy match against QName (or match all if empty)
		if nameQuery == "" {
			fl.Matches = append(fl.Matches, pickerMatch{item: it})
		} else if ok, sc := fuzzyScore(it.QName, nameQuery); ok {
			fl.Matches = append(fl.Matches, pickerMatch{item: it, score: sc})
		}
	}
	// Sort by score descending (insertion sort, small n)
	for i := 1; i < len(fl.Matches); i++ {
		for j := i; j > 0 && fl.Matches[j].score > fl.Matches[j-1].score; j-- {
			fl.Matches[j], fl.Matches[j-1] = fl.Matches[j-1], fl.Matches[j]
		}
	}
	if fl.Cursor >= len(fl.Matches) {
		fl.Cursor = max(0, len(fl.Matches)-1)
	}
	fl.Offset = 0
}

// MoveDown advances the cursor, wrapping to the top.
func (fl *FuzzyList) MoveDown() {
	if len(fl.Matches) == 0 {
		return
	}
	fl.Cursor++
	if fl.Cursor >= len(fl.Matches) {
		fl.Cursor = 0
		fl.Offset = 0
	} else if fl.Cursor >= fl.Offset+fl.MaxShow {
		fl.Offset = fl.Cursor - fl.MaxShow + 1
	}
}

// MoveUp moves the cursor up, wrapping to the bottom.
func (fl *FuzzyList) MoveUp() {
	if len(fl.Matches) == 0 {
		return
	}
	fl.Cursor--
	if fl.Cursor < 0 {
		fl.Cursor = len(fl.Matches) - 1
		fl.Offset = max(0, fl.Cursor-fl.MaxShow+1)
	} else if fl.Cursor < fl.Offset {
		fl.Offset = fl.Cursor
	}
}

// Selected returns the currently highlighted item, or an empty PickerItem.
func (fl FuzzyList) Selected() PickerItem {
	if len(fl.Matches) == 0 || fl.Cursor >= len(fl.Matches) {
		return PickerItem{}
	}
	return fl.Matches[fl.Cursor].item
}

// VisibleEnd returns the index just past the last visible match.
func (fl FuzzyList) VisibleEnd() int {
	return min(fl.Offset+fl.MaxShow, len(fl.Matches))
}
