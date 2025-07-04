package tui

import (
	"context"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
	"strings"
)

func ShowAddRelation(app *tview.Application) {
	form := tview.NewForm()
	form.
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
				ShowMessageAndReturnToMenu(i18n.T("invalid_format"))
				return
			}

			AsyncCallPages(app, i18n.T("adding_relation"), func() (string, string) {
				tuple := &v1.Relationship{
					Resource: &v1.ObjectReference{ObjectType: rp[0], ObjectId: rp[1]},
					Relation: rel,
					Subject:  &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: sp[0], ObjectId: sp[1]}},
				}
				_, err := client.Client.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
					Updates: []*v1.RelationshipUpdate{{Operation: v1.RelationshipUpdate_OPERATION_CREATE, Relationship: tuple}},
				})
				if err != nil {
					return i18n.T("error_adding_relation", err), i18n.T("error")
				}
				return i18n.T("relation_added_success"), i18n.T("success")
			})
		}).
		AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })

	form.SetBorder(true).SetTitle(i18n.T("add_relation")).SetTitleAlign(tview.AlignLeft)
	AddEscBack(form, "mainmenu")
	appPages.AddAndSwitchToPage("addrelation", form, true)
}
