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
	form := tview.NewForm()
	form = tview.NewForm().
		AddInputField(i18n.T("object_relations"), "", 30, nil, nil).
		AddButton(i18n.T("continue"), func() {
			id := form.GetFormItemByLabel(i18n.T("object_relations")).(*tview.InputField).GetText()
			showRelationsFor(app, "resource", id)
		}).
		AddButton(i18n.T("exit"), func() { app.SetRoot(BuildMainMenu(app), true) })

	form.SetBorder(true).SetTitle(i18n.T("object_relations")).SetTitleAlign(tview.AlignLeft)
	AddFormReturnESC(form, app, func() { app.SetRoot(BuildMainMenu(app), true) })
	app.SetRoot(form, true)
}

func ShowUserRelations(app *tview.Application) {
	form := tview.NewForm()
	form = tview.NewForm().
		AddInputField(i18n.T("user_relations"), "", 30, nil, nil).
		AddButton(i18n.T("continue"), func() {
			id := form.GetFormItemByLabel(i18n.T("user_relations")).(*tview.InputField).GetText()
			showRelationsFor(app, "subject", id)
		}).
		AddButton(i18n.T("exit"), func() { app.SetRoot(BuildMainMenu(app), true) })

	form.SetBorder(true).SetTitle(i18n.T("user_relations")).SetTitleAlign(tview.AlignLeft)
	AddFormReturnESC(form, app, func() { app.SetRoot(BuildMainMenu(app), true) })
	app.SetRoot(form, true)
}

func showRelationsFor(app *tview.Application, kind, identifier string) {
	parts := strings.SplitN(identifier, ":", 2)
	if len(parts) != 2 {
		ShowMessageAndReturnToMenu(app, "Invalid format – expected type:id")
		return
	}

	var filter *v1.RelationshipFilter
	if kind == "subject" {
		filter = &v1.RelationshipFilter{OptionalSubjectFilter: &v1.SubjectFilter{SubjectType: parts[0], OptionalSubjectId: parts[1]}}
	} else {
		filter = &v1.RelationshipFilter{ResourceType: parts[0], OptionalResourceId: parts[1]}
	}

	stream, err := client.Client.ReadRelationships(context.Background(), &v1.ReadRelationshipsRequest{RelationshipFilter: filter})
	tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	tv.SetBorder(true).SetTitle(fmt.Sprintf(" 🔍 Relations for %s ", identifier))

	if err != nil {
		tv.SetText(fmt.Sprintf("[red]Error: %v", err))
	} else {
		for {
			rel, err := stream.Recv()
			if err != nil {
				break
			}
			t := rel.GetRelationship()
			tv.Write([]byte(fmt.Sprintf("%s:%s#%s@%s:%s\n",
				t.GetResource().GetObjectType(), t.GetResource().GetObjectId(),
				t.GetRelation(),
				t.GetSubject().GetObject().GetObjectType(), t.GetSubject().GetObject().GetObjectId(),
			)))
		}
	}

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.SetRoot(BuildMainMenu(app), true)
			return nil
		}
		return event
	})
	app.SetRoot(tv, true)
}
