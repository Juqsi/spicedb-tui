package tui

import (
	"context"
	"fmt"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowAllTuples(app *tview.Application) {
	tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	tv.SetBorder(true).SetTitle(i18n.T("all_tuples"))

	schemaResp, err := client.Client.ReadSchema(context.Background(), nil)
	if err != nil {
		tv.SetText(fmt.Sprintf("[red]Error reading schema: %v", err))
		app.SetRoot(tv, true)
		return
	}

	var objectTypes []string
	for _, line := range strings.Split(schemaResp.SchemaText, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "definition ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				objectTypes = append(objectTypes, parts[1])
			}
		}
	}

	if len(objectTypes) == 0 {
		tv.SetText("[yellow]No object types found in schema.")
		app.SetRoot(tv, true)
		return
	}
	total := 0
	for _, objectType := range objectTypes {
		tv.Write([]byte(fmt.Sprintf("\n[::b][white]📦 %s[::-]\n", objectType)))
		stream, err := client.Client.ReadRelationships(context.Background(), &v1.ReadRelationshipsRequest{
			RelationshipFilter: &v1.RelationshipFilter{ResourceType: objectType},
		})
		if err != nil {
			tv.Write([]byte(fmt.Sprintf("  [red]Error fetching: %v\n", err)))
			continue
		}
		count := 0
		for {
			rel, err := stream.Recv()
			if err != nil {
				break
			}
			t := rel.GetRelationship()
			tv.Write([]byte(fmt.Sprintf("  %s:%s#%s@%s:%s\n",
				t.GetResource().GetObjectType(), t.GetResource().GetObjectId(),
				t.GetRelation(),
				t.GetSubject().GetObject().GetObjectType(), t.GetSubject().GetObject().GetObjectId(),
			)))
			count++
			total++
		}
		if count == 0 {
			tv.Write([]byte("  [gray](no tuples)\n"))
		}
	}

	if total == 0 {
		tv.Write([]byte("\n[yellow]⚠️ No tuples in the database."))
	} else {
		tv.Write([]byte(fmt.Sprintf("\n[green]✔ Loaded %d tuples.\n", total)))
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
