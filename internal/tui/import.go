package tui

import (
	"context"
	"strings"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowDataImport(app *tview.Application) {
	form := tview.NewForm()
	form = tview.NewForm().
		AddInputField("Demo tuple (type:id#rel@type:id)", "", 50, nil, nil).
		AddButton(i18n.T("continue"), func() {
			input := form.GetFormItemByLabel("Demo tuple (type:id#rel@type:id)").(*tview.InputField).GetText()
			parts := strings.Split(input, "#")
			if len(parts) != 2 {
				ShowMessageAndReturnToMenu(app, "Invalid format.")
				return
			}
			res := strings.Split(parts[0], ":")
			rest := strings.Split(parts[1], "@")
			if len(rest) != 2 || len(res) != 2 {
				ShowMessageAndReturnToMenu(app, "Invalid format.")
				return
			}
			rel := rest[0]
			sub := strings.Split(rest[1], ":")
			if len(sub) != 2 {
				ShowMessageAndReturnToMenu(app, "Invalid format.")
				return
			}
			tuple := &v1.Relationship{
				Resource: &v1.ObjectReference{ObjectType: res[0], ObjectId: res[1]},
				Relation: rel,
				Subject:  &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: sub[0], ObjectId: sub[1]}},
			}
			_, err := client.Client.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
				Updates: []*v1.RelationshipUpdate{{Operation: v1.RelationshipUpdate_OPERATION_CREATE, Relationship: tuple}},
			})
			if err != nil {
				ShowMessageAndReturnToMenu(app, "Import failed: %v", err)
			} else {
				ShowMessageAndReturnToMenu(app, "Import successful.")
			}
		}).
		AddButton(i18n.T("exit"), func() { app.SetRoot(BuildMainMenu(app), true) })

	form.SetBorder(true).SetTitle(i18n.T("data_import")).SetTitleAlign(tview.AlignLeft)
	AddFormReturnESC(form, app, func() { app.SetRoot(BuildMainMenu(app), true) })
	app.SetRoot(form, true)
}
