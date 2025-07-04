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
	form.AddInputField(i18n.T("upload_schema"), "", 200, nil, nil)
	form.AddButton(i18n.T("continue"), func() {
		schema := form.GetFormItemByLabel(i18n.T("upload_schema")).(*tview.InputField).GetText()
		AsyncCallPages(app, i18n.T("uploading_schema"), func() (string, string) {
			_, err := client.Client.WriteSchema(context.Background(), &v1.WriteSchemaRequest{Schema: schema})
			if err != nil {
				return i18n.T("error_writing_schema", err), i18n.T("error")
			}
			return i18n.T("schema_uploaded_success"), i18n.T("upload_schema")
		})
	})
	form.AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })
	form.SetBorder(true).SetTitle(i18n.T("upload_schema")).SetTitleAlign(tview.AlignLeft)
	AddEscBack(form, "mainmenu")
	appPages.AddAndSwitchToPage("writeschema", form, true)
}
