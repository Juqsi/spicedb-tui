package tui

import (
	"context"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

// subjectNodeRef is attached to tree nodes that represent a subject in an expanded leaf.
type subjectNodeRef struct {
	subject *v1.SubjectReference
	path    map[string]bool // object#relation keys of the ancestors, for cycle detection
	cycle   bool
	loading bool
	loaded  bool
}

func ShowExpandTree(app *tview.Application) {
	showExpandForm(app, "", "")
}

func showExpandForm(app *tview.Application, oldRes, oldPerm string) {
	form := tview.NewForm()
	form.
		AddInputField(i18n.T("resource"), oldRes, 30, nil, nil).
		AddInputField(i18n.T("permission_or_relation"), oldPerm, 20, nil, nil).
		AddButton(i18n.T("continue"), func() {
			res := form.GetFormItemByLabel(i18n.T("resource")).(*tview.InputField).GetText()
			perm := form.GetFormItemByLabel(i18n.T("permission_or_relation")).(*tview.InputField).GetText()
			obj, ok := parseObjectRef(res)
			if !ok || perm == "" {
				ShowMessageAndReturnToMenu(i18n.T("invalid_format"))
				return
			}
			openExpandTree(app, obj, perm, func() { showExpandForm(app, res, perm) })
		}).
		AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })
	form.SetBorder(true).SetTitle(i18n.T("expand_tree")).SetTitleAlign(tview.AlignLeft)
	AddEscBack(form, "mainmenu")
	appPages.AddAndSwitchToPage("expandform", form, true)
}

// openExpandTree expands obj#perm and shows the result as a tree. onBack is called on Esc.
func openExpandTree(app *tview.Application, obj *v1.ObjectReference, perm string, onBack func()) {
	showLoadingPage(i18n.T("expanding"))
	go func() {
		rsp, err := client.Client.ExpandPermissionTree(context.Background(), &v1.ExpandPermissionTreeRequest{
			Consistency: fullyConsistent(),
			Resource:    obj,
			Permission:  perm,
		})
		app.QueueUpdateDraw(func() {
			if err != nil {
				showTextResult("expanderror", i18n.T("error"), "[red]"+tview.Escape(i18n.T("error_expanding", err)), onBack)
				return
			}
			rootKey := formatObject(obj) + "#" + perm
			root := buildExpandNode(rsp.GetTreeRoot(), map[string]bool{rootKey: true}, "", 0)
			title := " 🌳 " + tview.Escape(rootKey) + " "
			showTreePage("expandtree", title, root, i18n.T("expand_hint"), onBack, func(n *tview.TreeNode) {
				onExpandSelect(app, n)
			})
		})
	}()
}

// buildExpandNode converts an expand-tree node into a tview tree node.
// Nodes at depth > 1 start collapsed.
func buildExpandNode(t *v1.PermissionRelationshipTree, path map[string]bool, role string, depth int) *tview.TreeNode {
	label := role
	if o := t.GetExpandedObject(); o != nil {
		label += tview.Escape(formatObject(o) + "#" + t.GetExpandedRelation())
	}
	node := tview.NewTreeNode("").SetExpanded(depth <= 1)

	switch tt := t.GetTreeType().(type) {
	case *v1.PermissionRelationshipTree_Intermediate:
		op := tt.Intermediate.GetOperation()
		node.SetText(label + "  " + opLabel(op))
		for i, c := range tt.Intermediate.GetChildren() {
			childRole := ""
			if op == v1.AlgebraicSubjectSet_OPERATION_EXCLUSION {
				if i == 0 {
					childRole = "[gray]" + i18n.T("base") + ":[-] "
				} else {
					childRole = "[red]" + i18n.T("minus") + ":[-] "
				}
			}
			node.AddChild(buildExpandNode(c, path, childRole, depth+1))
		}
	case *v1.PermissionRelationshipTree_Leaf:
		subjects := tt.Leaf.GetSubjects()
		node.SetText(label + "  [gray]" + i18n.T("n_subjects", len(subjects)) + "[-]")
		for _, s := range subjects {
			node.AddChild(newSubjectNode(s, path))
		}
	default:
		node.SetText(label)
	}
	return node
}

func newSubjectNode(s *v1.SubjectReference, path map[string]bool) *tview.TreeNode {
	ref := &subjectNodeRef{subject: s, path: path}
	text := tview.Escape(formatSubject(s))
	if s.GetOptionalRelation() != "" {
		if path[formatSubject(s)] {
			ref.cycle = true
			text += " [gray](" + i18n.T("cycle") + ")[-]"
		} else {
			text += " [gray]▸[-]"
		}
	}
	return tview.NewTreeNode(text).SetReference(ref).SetExpanded(false)
}

// onExpandSelect handles Enter on a node of the expand tree:
// subject sets are expanded lazily, concrete subjects open the tuple explorer,
// all other nodes are toggled.
func onExpandSelect(app *tview.Application, n *tview.TreeNode) {
	ref, ok := n.GetReference().(*subjectNodeRef)
	if !ok {
		n.SetExpanded(!n.IsExpanded())
		return
	}
	s := ref.subject
	if s.GetOptionalRelation() == "" {
		openTupleExplorer(app, explorerFilter{Subject: formatSubject(s)}, explorerFocusTable, true,
			func() { appPages.SwitchToPage("expandtree") })
		return
	}
	switch {
	case ref.cycle, ref.loading:
		return
	case ref.loaded:
		n.SetExpanded(!n.IsExpanded())
		return
	}

	ref.loading = true
	base := tview.Escape(formatSubject(s))
	n.SetText(base + " [gray]…[-]")
	go func() {
		rsp, err := client.Client.ExpandPermissionTree(context.Background(), &v1.ExpandPermissionTreeRequest{
			Consistency: fullyConsistent(),
			Resource:    s.GetObject(),
			Permission:  s.GetOptionalRelation(),
		})
		app.QueueUpdateDraw(func() {
			ref.loading = false
			if err != nil {
				n.SetText(base + " [red](" + tview.Escape(err.Error()) + ")[-]")
				return
			}
			ref.loaded = true
			newPath := make(map[string]bool, len(ref.path)+1)
			for k := range ref.path {
				newPath[k] = true
			}
			newPath[formatSubject(s)] = true
			sub := buildExpandNode(rsp.GetTreeRoot(), newPath, "", 1)
			n.SetText(sub.GetText())
			for _, c := range sub.GetChildren() {
				n.AddChild(c)
			}
			n.SetExpanded(true)
		})
	}()
}
