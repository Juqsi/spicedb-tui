package tui

import (
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/i18n"
)

// showTreePage displays a navigable tree with a hint line below it.
// Space toggles the current node; Enter calls onSelect (or toggles when onSelect is nil);
// Esc calls onEsc.
func showTreePage(pageName, title string, root *tview.TreeNode, hint string, onEsc func(), onSelect func(*tview.TreeNode)) *tview.TreeView {
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	tree.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignLeft)

	toggle := func(n *tview.TreeNode) {
		if n != nil {
			n.SetExpanded(!n.IsExpanded())
		}
	}

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEsc:
			onEsc()
			return nil
		case event.Key() == tcell.KeyRune && event.Rune() == ' ':
			toggle(tree.GetCurrentNode())
			return nil
		case event.Key() == tcell.KeyEnter:
			if n := tree.GetCurrentNode(); n != nil {
				if onSelect != nil {
					onSelect(n)
				} else {
					toggle(n)
				}
			}
			return nil
		}
		return event
	})

	hintView := tview.NewTextView().SetDynamicColors(true).SetText("[gray]" + hint)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tree, 0, 1, true).
		AddItem(hintView, 1, 0, false)
	appPages.AddAndSwitchToPage(pageName, layout, true)
	return tree
}

// opLabel renders an algebraic set operation with color.
func opLabel(op v1.AlgebraicSubjectSet_Operation) string {
	switch op {
	case v1.AlgebraicSubjectSet_OPERATION_UNION:
		return "[green]∪ " + i18n.T("union") + "[-]"
	case v1.AlgebraicSubjectSet_OPERATION_INTERSECTION:
		return "[blue]∩ " + i18n.T("intersection") + "[-]"
	case v1.AlgebraicSubjectSet_OPERATION_EXCLUSION:
		return "[red]− " + i18n.T("exclusion") + "[-]"
	}
	return "[gray]?[-]"
}

// resultLabel renders a debug-trace result as a colored symbol.
func resultLabel(r v1.CheckDebugTrace_Permissionship) string {
	switch r {
	case v1.CheckDebugTrace_PERMISSIONSHIP_HAS_PERMISSION:
		return "[green]✔[-]"
	case v1.CheckDebugTrace_PERMISSIONSHIP_NO_PERMISSION:
		return "[red]✘[-]"
	case v1.CheckDebugTrace_PERMISSIONSHIP_CONDITIONAL_PERMISSION:
		return "[yellow]?[-]"
	}
	return "[gray]·[-]"
}
