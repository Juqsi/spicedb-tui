package tui

import (
	"context"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
)

func ShowSchema(app *tview.Application) {
	AsyncCall(app, "[yellow]loading...", func() (string, string) {
		rsp, err := client.Client.ReadSchema(context.Background(), nil)
		if err != nil {
			return "[red]Error: " + err.Error(), "Error"
		}
		return rsp.SchemaText, "SpiceDB Schema"
	})
}
