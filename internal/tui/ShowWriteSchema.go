package tui

import (
	"context"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowWriteSchema(app *tview.Application) {
	form := tview.NewForm()
	form = tview.NewForm().
		AddInputField(i18n.T("upload_schema"), "", 60, nil, nil).
		AddButton(i18n.T("continue"), func() {
			schema := form.GetFormItemByLabel(i18n.T("upload_schema")).(*tview.InputField).GetText()
			_, err := client.Client.WriteSchema(context.Background(), &v1.WriteSchemaRequest{Schema: schema})
			if err != nil {
				ShowMessageAndReturnToMenu(app, "Error writing schema: %v", err)
			} else {
				ShowMessageAndReturnToMenu(app, "Schema uploaded successfully.")
			}
		}).
		AddButton(i18n.T("exit"), func() { app.SetRoot(BuildMainMenu(app), true) })

	form.SetBorder(true).SetTitle(i18n.T("upload_schema")).SetTitleAlign(tview.AlignLeft)
	AddFormReturnESC(form, app, func() { app.SetRoot(BuildMainMenu(app), true) })
	app.SetRoot(form, true)
}
