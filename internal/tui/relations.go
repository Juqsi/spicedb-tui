package tui

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowObjectRelations(app *tview.Application) {
	showRelationsForm(app, "object")
}

func ShowUserRelations(app *tview.Application) {
	showRelationsForm(app, "user")
}

func showRelationsForm(app *tview.Application, kind string) {
	label := i18n.T("object_relations")
	if kind == "user" {
		label = i18n.T("user_relations")
	}
	form := tview.NewForm()
	form.AddInputField(label, "", 30, nil, nil)
	form.AddButton(i18n.T("continue"), func() {
		id := form.GetFormItemByLabel(label).(*tview.InputField).GetText()
		showRelationsForPages(app, kind, id)
	})
	form.AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })
	form.SetBorder(true).SetTitle(label).SetTitleAlign(tview.AlignLeft)
	AddEscBack(form, "mainmenu")
	appPages.AddAndSwitchToPage(kind+"relform", form, true)
}

func showRelationsForPages(app *tview.Application, kind, identifier string) {
	parts := strings.SplitN(identifier, ":", 2)
	if len(parts) != 2 {
		ShowMessageAndReturnToMenu(i18n.T("invalid_format_typeid"))
		return
	}
	var filter *v1.RelationshipFilter
	if kind == "user" {
		filter = &v1.RelationshipFilter{OptionalSubjectFilter: &v1.SubjectFilter{SubjectType: parts[0], OptionalSubjectId: parts[1]}}
	} else {
		filter = &v1.RelationshipFilter{ResourceType: parts[0], OptionalResourceId: parts[1]}
	}
	loading := tview.NewTextView().SetText(i18n.T("loading_relations")).SetBorder(true).SetTitle(i18n.T("loading"))
	appPages.AddAndSwitchToPage(kind+"relload", loading, true)
	go func() {
		stream, err := client.Client.ReadRelationships(context.Background(), &v1.ReadRelationshipsRequest{RelationshipFilter: filter})
		var out strings.Builder
		if err != nil {
			out.WriteString(fmt.Sprintf("[red]"+i18n.T("error_reading_relations"), err))
		} else {
			for {
				rel, err := stream.Recv()
				if err != nil {
					break
				}
				t := rel.GetRelationship()
				out.WriteString(fmt.Sprintf("%s:%s#%s@%s:%s\n",
					t.GetResource().GetObjectType(), t.GetResource().GetObjectId(),
					t.GetRelation(),
					t.GetSubject().GetObject().GetObjectType(), t.GetSubject().GetObject().GetObjectId(),
				))
			}
		}
		app.QueueUpdateDraw(func() {
			label := i18n.T("object_relations")
			if kind == "user" {
				label = i18n.T("user_relations")
			}
			resultView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
			resultView.SetBorder(true).SetTitle(fmt.Sprintf(" 🔍 %s: %s ", label, identifier))
			resultView.SetText(out.String())
			resultView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					form := tview.NewForm()
					form.AddInputField(label, identifier, 30, nil, nil)
					form.AddButton(i18n.T("continue"), func() {
						id := form.GetFormItemByLabel(label).(*tview.InputField).GetText()
						showRelationsForPages(app, kind, id)
					})
					form.AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })
					form.SetBorder(true).SetTitle(label).SetTitleAlign(tview.AlignLeft)
					AddEscBack(form, "mainmenu")
					appPages.AddAndSwitchToPage(kind+"relform", form, true)
					return nil
				}
				return event
			})
			appPages.AddAndSwitchToPage(kind+"relresult", resultView, true)
		})
	}()
}
