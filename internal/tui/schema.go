package tui

import (
	"context"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowSchema(app *tview.Application) {
	AsyncCallPages(app, i18n.T("loading"), func() (string, string) {
		rsp, err := client.Client.ReadSchema(context.Background(), nil)
		if err != nil {
			return i18n.T("error_loading_schema", err), i18n.T("error")
		}
		return rsp.SchemaText, i18n.T("schema_title")
	})
}
