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
	form.
		AddInputField(i18n.T("demo_tuple_format"), "", 50, nil, nil).
		AddButton(i18n.T("continue"), func() {
			input := form.GetFormItemByLabel(i18n.T("demo_tuple_format")).(*tview.InputField).GetText()
			parts := strings.Split(input, "#")
			if len(parts) != 2 {
				ShowMessageAndReturnToMenu(i18n.T("invalid_tuple_format"))
				return
			}
			res := strings.Split(parts[0], ":")
			rest := strings.Split(parts[1], "@")
			if len(rest) != 2 || len(res) != 2 {
				ShowMessageAndReturnToMenu(i18n.T("invalid_tuple_format"))
				return
			}
			rel := rest[0]
			sub := strings.Split(rest[1], ":")
			if len(sub) != 2 {
				ShowMessageAndReturnToMenu(i18n.T("invalid_tuple_format"))
				return
			}
			AsyncCallPages(app, i18n.T("importing_tuple"), func() (string, string) {
				tuple := &v1.Relationship{
					Resource: &v1.ObjectReference{ObjectType: res[0], ObjectId: res[1]},
					Relation: rel,
					Subject:  &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: sub[0], ObjectId: sub[1]}},
				}
				_, err := client.Client.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
					Updates: []*v1.RelationshipUpdate{{Operation: v1.RelationshipUpdate_OPERATION_CREATE, Relationship: tuple}},
				})
				if err != nil {
					return i18n.T("import_failed", err), i18n.T("data_import")
				} else {
					return i18n.T("import_successful"), i18n.T("data_import")
				}
			})
		}).
		AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })
	form.SetBorder(true).SetTitle(i18n.T("data_import")).SetTitleAlign(tview.AlignLeft)
	AddEscBack(form, "mainmenu")
	appPages.AddAndSwitchToPage("dataimport", form, true)
}
