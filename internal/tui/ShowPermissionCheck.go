package tui

import (
	"context"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
	"strings"
)

func ShowPermissionCheck(app *tview.Application) {
	form := tview.NewForm()
	form = tview.NewForm().
		AddInputField("Resource (type:id)", "", 30, nil, nil).
		AddInputField("Permission", "", 20, nil, nil).
		AddInputField("Subject (type:id)", "", 30, nil, nil).
		AddButton(i18n.T("continue"), func() {
			res := strings.Split(form.GetFormItemByLabel("Resource (type:id)").(*tview.InputField).GetText(), ":")
			sub := strings.Split(form.GetFormItemByLabel("Subject (type:id)").(*tview.InputField).GetText(), ":")
			perm := form.GetFormItemByLabel("Permission").(*tview.InputField).GetText()

			if len(res) != 2 || len(sub) != 2 {
				ShowMessageAndReturnToMenu(app, "Invalid format (type:id)")
				return
			}

			rsp, err := client.Client.CheckPermission(context.Background(), &v1.CheckPermissionRequest{
				Resource:   &v1.ObjectReference{ObjectType: res[0], ObjectId: res[1]},
				Permission: perm,
				Subject:    &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: sub[0], ObjectId: sub[1]}},
			})
			if err != nil {
				ShowMessageAndReturnToMenu(app, "Error: %v", err)
				return
			}
			result := "NOT allowed"
			if rsp.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
				result = "allowed"
			}
			ShowMessageAndReturnToMenu(app, "Result: access is %s", result)
		}).
		AddButton(i18n.T("exit"), func() { app.SetRoot(BuildMainMenu(app), true) })

	form.SetBorder(true).SetTitle(i18n.T("permission_check")).SetTitleAlign(tview.AlignLeft)
	AddFormReturnESC(form, app, func() { app.SetRoot(BuildMainMenu(app), true) })
	app.SetRoot(form, true)
}
