// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"strconv"
	"strings"
)

// MenuHint represents keyboard mnemonic.
type MenuHint struct {
	Mnemonic    string
	Description string
	Visible     bool
}

// IsBlank checks if menu hint is a place holder.
func (m MenuHint) IsBlank() bool {
	return m.Mnemonic == "" && m.Description == "" && !m.Visible
}

// String returns a string representation.
func (m MenuHint) String() string {
	return m.Mnemonic
}

// MenuHints represents a collection of hints.
type MenuHints []MenuHint

// Len returns the hints length.
func (h MenuHints) Len() int {
	return len(h)
}

// Swap swaps to elements.
func (h MenuHints) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Mnemonic groups, used to lay hints out into distinct menu columns.
const (
	GroupContext = iota // S-1..S-9 (context favorites)
	GroupNamespace      // 0..9 (namespace favorites)
	GroupPlain          // everything else
)

// MnemonicGroup buckets a mnemonic for sorting/column layout purposes.
func MnemonicGroup(mnemonic string) int {
	if rest, ok := strings.CutPrefix(mnemonic, "S-"); ok {
		if _, err := strconv.Atoi(rest); err == nil {
			return GroupContext
		}
	}
	if _, err := strconv.Atoi(mnemonic); err == nil {
		return GroupNamespace
	}
	return GroupPlain
}

// priorityRank returns a sort rank for mnemonics that should be grouped and
// ordered ahead of plain alphabetic hotkeys: S-digits (S-1..S-9, context
// favorites) rank first, followed by plain digits (0..9, namespace
// favorites) as a second contiguous group.
func priorityRank(mnemonic string) (int, bool) {
	if rest, ok := strings.CutPrefix(mnemonic, "S-"); ok {
		if n, err := strconv.Atoi(rest); err == nil {
			return n, true
		}
	}
	if n, err := strconv.Atoi(mnemonic); err == nil {
		return 100 + n, true
	}
	return 0, false
}

// Less returns true if first hint is less than second.
func (h MenuHints) Less(i, j int) bool {
	n, oki := priorityRank(h[i].Mnemonic)
	m, okj := priorityRank(h[j].Mnemonic)
	if oki && okj {
		return n < m
	}
	if oki && !okj {
		return true
	}
	if !oki && okj {
		return false
	}
	return h[i].Description < h[j].Description
}
