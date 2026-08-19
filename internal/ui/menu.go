// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	menuIndexFmt = " [key:-:b]<%s> [fg:-:fgstyle]%s "
	maxRows      = 6
)

// Menu presents menu options.
type Menu struct {
	*tview.Table

	styles *config.Styles
}

// NewMenu returns a new menu.
func NewMenu(styles *config.Styles) *Menu {
	m := Menu{
		Table:  tview.NewTable(),
		styles: styles,
	}
	m.SetBackgroundColor(styles.BgColor())
	styles.AddListener(&m)

	return &m
}

// StylesChanged notifies skin changed.
func (m *Menu) StylesChanged(s *config.Styles) {
	m.styles = s
	m.SetBackgroundColor(s.BgColor())
	for row := range m.GetRowCount() {
		for col := range m.GetColumnCount() {
			if c := m.GetCell(row, col); c != nil {
				c.BackgroundColor = s.BgColor()
			}
		}
	}
}

// StackPushed notifies a component was added.
func (m *Menu) StackPushed(c model.Component) {
	m.HydrateMenu(c.Hints())
}

// StackPopped notifies a component was removed.
func (m *Menu) StackPopped(_, top model.Component) {
	if top != nil {
		m.HydrateMenu(top.Hints())
	} else {
		m.Clear()
	}
}

// StackTop notifies the top component.
func (m *Menu) StackTop(t model.Component) {
	m.HydrateMenu(t.Hints())
}

// HydrateMenu populate menu ui from hints.
func (m *Menu) HydrateMenu(hh model.MenuHints) {
	m.Clear()
	sort.Sort(hh)

	table := make([]model.MenuHints, maxRows+1)
	colCount := (len(hh) / maxRows) + 1 + m.priorityGroupCount(hh)
	for row := range maxRows {
		table[row] = make(model.MenuHints, colCount)
	}
	t := m.buildMenuTable(hh, table, colCount)

	for row := range t {
		for col := range len(t[row]) {
			c := tview.NewTableCell(t[row][col])
			if t[row][col] == "" {
				c = tview.NewTableCell("")
			}
			c.SetBackgroundColor(m.styles.BgColor())
			m.SetCell(row, col, c)
		}
	}
}

// priorityGroupCount returns how many distinct priority groups (context
// favorites, namespace favorites) have at least one visible hint. Each such
// group gets forced onto its own dedicated column(s) in buildMenuTable, so
// the column budget needs a slot reserved per group present.
func (*Menu) priorityGroupCount(hh model.MenuHints) int {
	seen := make(map[int]bool, 2)
	for _, h := range hh {
		if !h.Visible {
			continue
		}
		if g := model.MnemonicGroup(h.Mnemonic); g != model.GroupPlain {
			seen[g] = true
		}
	}
	return len(seen)
}

func (m *Menu) buildMenuTable(hh model.MenuHints, table []model.MenuHints, colCount int) [][]string {
	var row, col int
	prevGroup := -1
	maxKeys := make([]int, colCount)
	for _, h := range hh {
		if !h.Visible {
			continue
		}

		group := model.MnemonicGroup(h.Mnemonic)
		if prevGroup != -1 && group != prevGroup {
			row, col = 0, col+1
			if table[0][0].IsBlank() {
				col = 0
			}
		}
		prevGroup = group

		if maxKeys[col] < len(h.Mnemonic) {
			maxKeys[col] = len(h.Mnemonic)
		}
		table[row][col] = h
		row++
		if row >= maxRows {
			row, col = 0, col+1
		}
	}

	out := make([][]string, len(table))
	for r := range out {
		out[r] = make([]string, len(table[r]))
	}
	m.layout(table, maxKeys, out)

	return out
}

func (m *Menu) layout(table []model.MenuHints, mm []int, out [][]string) {
	for r := range table {
		for c := range table[r] {
			out[r][c] = m.formatMenu(table[r][c], mm[c])
		}
	}
}

func (m *Menu) formatMenu(h model.MenuHint, size int) string {
	if h.Mnemonic == "" || h.Description == "" {
		return ""
	}
	styles := m.styles.Frame()
	switch model.MnemonicGroup(h.Mnemonic) {
	case model.GroupNamespace, model.GroupContext:
		return formatNSMenu(h.Mnemonic, h.Description, &styles)
	default:
		return formatPlainMenu(h, size, &styles)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func keyConv(s string) string {
	if s == "" || !strings.Contains(s, "alt") {
		return s
	}
	if runtime.GOOS != "darwin" {
		return s
	}

	return strings.Replace(s, "alt", "opt", 1)
}

// Truncate a string to the given l and suffix ellipsis if needed.
func Truncate(str string, width int) string {
	return runewidth.Truncate(str, width, string(tview.SemigraphicsHorizontalEllipsis))
}

func ToMnemonic(s string) string {
	if s == "" {
		return s
	}

	return "<" + keyConv(strings.ToLower(s)) + ">"
}

func formatNSMenu(key, name string, styles *config.Frame) string {
	fmat := strings.Replace(menuIndexFmt, "[key", "["+styles.Menu.NumKeyColor.String(), 1)
	fmat = strings.ReplaceAll(fmat, ":bg:", ":"+styles.Title.BgColor.String()+":")
	fmat = strings.Replace(fmat, "[fg", "["+styles.Menu.FgColor.String(), 1)
	fmat = strings.Replace(fmat, "fgstyle]", styles.Menu.FgStyle.ToShortString()+"]", 1)

	return fmt.Sprintf(fmat, key, name)
}

func formatPlainMenu(h model.MenuHint, size int, styles *config.Frame) string {
	menuFmt := " [key:-:b]%-" + strconv.Itoa(size+2) + "s [fg:-:fgstyle]%s "
	fmat := strings.Replace(menuFmt, "[key", "["+styles.Menu.KeyColor.String(), 1)
	fmat = strings.Replace(fmat, "[fg", "["+styles.Menu.FgColor.String(), 1)
	fmat = strings.ReplaceAll(fmat, ":bg:", ":"+styles.Title.BgColor.String()+":")
	fmat = strings.Replace(fmat, "fgstyle]", styles.Menu.FgStyle.ToShortString()+"]", 1)

	return fmt.Sprintf(fmat, ToMnemonic(h.Mnemonic), h.Description)
}
