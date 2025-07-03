package tui

import (
	"context"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowSchema(app *tview.Application) {
	tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	tv.SetBorder(true).SetTitle(i18n.T("show_schema"))

	AsyncCall(app, "[yellow] loading...", func() string {
		rsp, err := client.Client.ReadSchema(context.Background(), nil)
		if err != nil {
			return "[red]" + err.Error()
		} else {
			return "[red]" + rsp.SchemaText
		}
	})

	app.SetRoot(tv, true)
}
