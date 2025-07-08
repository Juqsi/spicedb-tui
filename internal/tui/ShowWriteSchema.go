package tui

import (
	"context"
	"os"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowWriteSchema(app *tview.Application) {
	textArea := tview.NewTextArea()
	textArea.
		SetBorder(true).
		SetTitle(i18n.T("upload_schema"))
	textArea.
		SetPlaceholder(i18n.T("schema_placeholder"))

	var filePath string

	form := tview.NewForm().
		AddInputField(i18n.T("schema_file_path"), "", 80, nil, func(s string) {
			filePath = s
		}).
		AddButton(i18n.T("continue"), func() {
			var schema string
			if filePath != "" {
				data, err := os.ReadFile(filePath)
				if err != nil {
					ShowMessageAndReturnToMenu(i18n.T("error_reading_file", err))
					return
				}
				schema = string(data)
			} else {
				schema = textArea.GetText()
			}
			AsyncCallPages(app, i18n.T("uploading_schema"), func() (string, string) {
				_, err := client.Client.WriteSchema(context.Background(), &v1.WriteSchemaRequest{Schema: schema})
				if err != nil {
					return i18n.T("error_writing_schema", err), i18n.T("error")
				}
				return i18n.T("schema_uploaded_success"), i18n.T("upload_schema")
			})
		}).
		AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })

	form.SetBorder(false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textArea, 0, 3, true).
		AddItem(form, 0, 1, false)

	layout.SetBorder(true).SetTitle(i18n.T("upload_schema")).SetTitleAlign(tview.AlignLeft)

	textArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(form)
			return nil
		case tcell.KeyEsc:
			appPages.SwitchToPage("mainmenu")
			return nil
		}
		return event
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBacktab:
			app.SetFocus(textArea)
			return nil
		case tcell.KeyEsc:
			appPages.SwitchToPage("mainmenu")
			return nil
		}
		return event
	})

	AddEscBack(layout, "mainmenu")
	appPages.AddAndSwitchToPage("writeschema", layout, true)
}
