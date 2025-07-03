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

func ShowAddRelation(app *tview.Application) {
	form := tview.NewForm()
	form = tview.NewForm().
		AddInputField("Resource (type:id)", "", 30, nil, nil).
		AddInputField("Relation", "", 20, nil, nil).
		AddInputField("Subject (type:id)", "", 30, nil, nil).
		AddButton(i18n.T("continue"), func() {
			res := form.GetFormItemByLabel("Resource (type:id)").(*tview.InputField).GetText()
			rel := form.GetFormItemByLabel("Relation").(*tview.InputField).GetText()
			sub := form.GetFormItemByLabel("Subject (type:id)").(*tview.InputField).GetText()

			rp := strings.SplitN(res, ":", 2)
			sp := strings.SplitN(sub, ":", 2)
			if len(rp) != 2 || len(sp) != 2 {
				ShowMessageAndReturnToMenu(app, "Invalid format (type:id)")
				return
			}

			AsyncCall(app, "[yellow]"+i18n.T("adding_relation"), func() string {
				tuple := &v1.Relationship{
					Resource: &v1.ObjectReference{ObjectType: rp[0], ObjectId: rp[1]},
					Relation: rel,
					Subject:  &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: sp[0], ObjectId: sp[1]}},
				}
				_, err := client.Client.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
					Updates: []*v1.RelationshipUpdate{{Operation: v1.RelationshipUpdate_OPERATION_CREATE, Relationship: tuple}},
				})
				if err != nil {
					return fmt.Sprintf("Error adding relation: %v", err)
				}
				return "Relation added successfully."
			})
		}).
		AddButton(i18n.T("exit"), func() { app.SetRoot(BuildMainMenu(app), true) })

	form.SetBorder(true).SetTitle(i18n.T("add_relation")).SetTitleAlign(tview.AlignLeft)
	AddFormReturnESC(form, app, func() { app.SetRoot(BuildMainMenu(app), true) })
	app.SetRoot(form, true)
}
