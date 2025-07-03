package tui

import (
	"context"
	"fmt"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
	"strings"
)

func ShowDeleteRelation(app *tview.Application) {
	form := tview.NewForm()
	form = tview.NewForm().
		AddInputField(i18n.T("resource"), "", 30, nil, nil).
		AddInputField(i18n.T("relation"), "", 20, nil, nil).
		AddInputField(i18n.T("subject"), "", 30, nil, nil).
		AddButton(i18n.T("continue"), func() {
			res := form.GetFormItemByLabel(i18n.T("resource")).(*tview.InputField).GetText()
			rel := form.GetFormItemByLabel(i18n.T("relation")).(*tview.InputField).GetText()
			sub := form.GetFormItemByLabel(i18n.T("subject")).(*tview.InputField).GetText()

			rp := strings.SplitN(res, ":", 2)
			sp := strings.SplitN(sub, ":", 2)
			if len(rp) != 2 || len(sp) != 2 {
				ShowMessageAndReturnToMenu(app, i18n.T("invalid_format"))
				return
			}

			tuple := &v1.Relationship{
				Resource: &v1.ObjectReference{ObjectType: rp[0], ObjectId: rp[1]},
				Relation: rel,
				Subject:  &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: sp[0], ObjectId: sp[1]}},
			}

			AsyncCall(app, i18n.T("deleting_relation"), func() (string, string) {
				_, err := client.Client.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
					Updates: []*v1.RelationshipUpdate{{
						Operation:    v1.RelationshipUpdate_OPERATION_DELETE,
						Relationship: tuple,
					}},
				})
				if err != nil {
					return fmt.Sprintf(i18n.T("error_deleting_relation"), err), "Error"
				}
				return fmt.Sprintf(i18n.T("relation_deleted_success")), "Delete Relation"
			})
		}).
		AddButton(i18n.T("exit"), func() { app.SetRoot(BuildMainMenu(app), true) })

	form.SetBorder(true).
		SetTitle(i18n.T("delete_relation")).
		SetTitleAlign(tview.AlignLeft)
	AddFormReturnESC(form, app, func() { app.SetRoot(BuildMainMenu(app), true) })
	app.SetRoot(form, true)
}
