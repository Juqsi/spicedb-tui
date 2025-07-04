package tui

import (
	"context"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
	"strings"
)

func ShowPermissionCheck(app *tview.Application) {
	showPermissionCheckForm(app, "", "", "")
}

func showPermissionCheckForm(app *tview.Application, oldRes, oldPerm, oldSub string) {
	form := tview.NewForm()
	form.
		AddInputField(i18n.T("resource"), oldRes, 30, nil, nil).
		AddInputField(i18n.T("permission"), oldPerm, 20, nil, nil).
		AddInputField(i18n.T("subject"), oldSub, 30, nil, nil).
		AddButton(i18n.T("continue"), func() {
			resource := form.GetFormItemByLabel(i18n.T("resource")).(*tview.InputField).GetText()
			permission := form.GetFormItemByLabel(i18n.T("permission")).(*tview.InputField).GetText()
			subject := form.GetFormItemByLabel(i18n.T("subject")).(*tview.InputField).GetText()
			AsyncCallPagesCustomBack(app, i18n.T("loading"), func() (string, string) {
				res := strings.Split(resource, ":")
				sub := strings.Split(subject, ":")
				if len(res) != 2 || len(sub) != 2 {
					return i18n.T("invalid_format"), i18n.T("error")
				}
				rsp, err := client.Client.CheckPermission(context.Background(), &v1.CheckPermissionRequest{
					Resource:   &v1.ObjectReference{ObjectType: res[0], ObjectId: res[1]},
					Permission: permission,
					Subject:    &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: sub[0], ObjectId: sub[1]}},
				})
				if err != nil {
					return i18n.T("error_check_permission", err), i18n.T("error")
				}
				result := "[red]" + i18n.T("not_allowed")
				if rsp.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
					result = "[green]" + i18n.T("allowed")
				}
				return i18n.T("check_result", result), i18n.T("permission_check_title", resource, permission, subject)
			}, func() {
				showPermissionCheckForm(app, resource, permission, subject)
			})
		}).
		AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })
	form.SetBorder(true).SetTitle(i18n.T("permission_check")).SetTitleAlign(tview.AlignLeft)
	AddEscBack(form, "mainmenu")
	appPages.AddAndSwitchToPage("permcheckform", form, true)
}

func AsyncCallPagesCustomBack(app *tview.Application, loadingText string, fn func() (result string, title string), backFunc func()) {
	loadingView := tview.NewTextView().
		SetText(loadingText).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle("⏳ " + i18n.T("loading"))
	appPages.AddAndSwitchToPage("loading", loadingView, true)
	go func() {
		resultText, resultTitle := fn()
		app.QueueUpdateDraw(func() {
			tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
			tv.SetBorder(true).
				SetTitle(resultTitle)
			tv.SetText(resultText)
			tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					backFunc()
					return nil
				}
				return event
			})
			appPages.AddAndSwitchToPage("permcheckresult", tv, true)
		})
	}()
}
