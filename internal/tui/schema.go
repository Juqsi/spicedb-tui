package tui

import (
	"context"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowSchema(app *tview.Application) {
	tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	tv.SetBorder(true).SetTitle(i18n.T("show_schema"))

	rsp, err := client.Client.ReadSchema(context.Background(), nil)
	if err != nil {
		tv.SetText("[red]" + err.Error())
	} else {
		tv.SetText(rsp.SchemaText)
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
