package tui

import (
	"context"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowPermissionCheck(app *tview.Application) {
	showPermissionCheckForm(app, "", "", "", true)
}

func showPermissionCheckForm(app *tview.Application, oldRes, oldPerm, oldSub string, oldExplain bool) {
	form := tview.NewForm()
	form.
		AddInputField(i18n.T("resource"), oldRes, 30, nil, nil).
		AddInputField(i18n.T("permission"), oldPerm, 20, nil, nil).
		AddInputField(i18n.T("subject"), oldSub, 30, nil, nil).
		AddCheckbox(i18n.T("explain_trace"), oldExplain, nil).
		AddButton(i18n.T("continue"), func() {
			resource := form.GetFormItemByLabel(i18n.T("resource")).(*tview.InputField).GetText()
			permission := form.GetFormItemByLabel(i18n.T("permission")).(*tview.InputField).GetText()
			subject := form.GetFormItemByLabel(i18n.T("subject")).(*tview.InputField).GetText()
			explain := form.GetFormItemByLabel(i18n.T("explain_trace")).(*tview.Checkbox).IsChecked()
			back := func() { showPermissionCheckForm(app, resource, permission, subject, explain) }
			runPermissionCheck(app, resource, permission, subject, explain, back)
		}).
		AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })
	form.SetBorder(true).SetTitle(i18n.T("permission_check")).SetTitleAlign(tview.AlignLeft)
	AddEscBack(form, "mainmenu")
	appPages.AddAndSwitchToPage("permcheckform", form, true)
}

func runPermissionCheck(app *tview.Application, resource, permission, subject string, explain bool, back func()) {
	resObj, okRes := parseObjectRef(resource)
	subRef, okSub := parseSubjectRef(subject)
	if !okRes || !okSub {
		showTextResult("permcheckresult", i18n.T("error"), i18n.T("invalid_format"), back)
		return
	}
	title := tview.Escape(i18n.T("permission_check_title", resource, permission, subject))

	showLoadingPage(i18n.T("loading"))
	go func() {
		rsp, err := client.Client.CheckPermission(context.Background(), &v1.CheckPermissionRequest{
			Consistency: fullyConsistent(),
			Resource:    resObj,
			Permission:  permission,
			Subject:     subRef,
			WithTracing: explain,
		})
		app.QueueUpdateDraw(func() {
			if err != nil {
				showTextResult("permcheckresult", i18n.T("error"), "[red]"+tview.Escape(i18n.T("error_check_permission", err)), back)
				return
			}
			verdict := i18n.T("check_result", permissionshipText(rsp.GetPermissionship()))
			if trace := rsp.GetDebugTrace().GetCheck(); explain && trace != nil {
				denied := rsp.GetPermissionship() == v1.CheckPermissionResponse_PERMISSIONSHIP_NO_PERMISSION
				root := buildTraceNode(trace, 0, denied)
				showTreePage("permchecktree", " "+verdict+"  "+title+" ", root, i18n.T("trace_hint"), back, nil)
				return
			}
			text := verdict
			if explain {
				text += "\n\n[gray]" + i18n.T("no_trace") + "[-]"
			}
			showTextResult("permcheckresult", title, text, back)
		})
	}()
}

func permissionshipText(p v1.CheckPermissionResponse_Permissionship) string {
	switch p {
	case v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION:
		return "[green]" + i18n.T("allowed") + "[-]"
	case v1.CheckPermissionResponse_PERMISSIONSHIP_CONDITIONAL_PERMISSION:
		return "[yellow]" + i18n.T("conditional_allowed") + "[-]"
	}
	return "[red]" + i18n.T("not_allowed") + "[-]"
}

// buildTraceNode converts a check debug trace into a tree. Granted (or conditional)
// branches are expanded; for a denied check everything up to depth 3 is expanded.
func buildTraceNode(t *v1.CheckDebugTrace, depth int, denied bool) *tview.TreeNode {
	label := resultLabel(t.GetResult()) + "  " + tview.Escape(formatObject(t.GetResource())+"#"+t.GetPermission())
	switch t.GetPermissionType() {
	case v1.CheckDebugTrace_PERMISSION_TYPE_RELATION:
		label += " [gray](" + i18n.T("relation_kind") + ")[-]"
	case v1.CheckDebugTrace_PERMISSION_TYPE_PERMISSION:
		label += " [gray](" + i18n.T("permission_kind") + ")[-]"
	}
	if s := t.GetSubject(); s != nil {
		label += " @ " + tview.Escape(formatSubject(s))
	}
	if d := t.GetDuration(); d != nil {
		label += " [gray]" + d.AsDuration().Round(time.Microsecond).String() + "[-]"
	}
	if t.GetWasCachedResult() {
		label += " [gray](" + i18n.T("cached") + ")[-]"
	}

	r := t.GetResult()
	granted := r == v1.CheckDebugTrace_PERMISSIONSHIP_HAS_PERMISSION || r == v1.CheckDebugTrace_PERMISSIONSHIP_CONDITIONAL_PERMISSION
	node := tview.NewTreeNode(label).SetExpanded(granted || (denied && depth < 3))
	for _, c := range t.GetSubProblems().GetTraces() {
		node.AddChild(buildTraceNode(c, depth+1, denied))
	}
	return node
}
