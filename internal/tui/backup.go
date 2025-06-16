package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowBackupCreate(app *tview.Application) {
	tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	tv.SetBorder(true).SetTitle(i18n.T("backup_create"))

	schemaResp, err := client.Client.ReadSchema(context.Background(), &v1.ReadSchemaRequest{})
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

	var backupData []string
	for _, objType := range objectTypes {
		stream, err := client.Client.ReadRelationships(context.Background(), &v1.ReadRelationshipsRequest{
			RelationshipFilter: &v1.RelationshipFilter{ResourceType: objType},
		})
		if err != nil {
			tv.Write([]byte(fmt.Sprintf("[red]Error for %s: %v\n", objType, err)))
			continue
		}
		for {
			rel, err := stream.Recv()
			if err != nil {
				break
			}
			bs, _ := json.Marshal(rel.GetRelationship())
			backupData = append(backupData, string(bs))
		}
	}

	_ = os.WriteFile("spicedb-backup.json", []byte(strings.Join(backupData, "\n")), 0644)

	if len(backupData) == 0 {
		tv.SetText("[yellow]No tuples to backup.")
	} else {
		tv.SetText(strings.Join(backupData, "\n"))
	}

	tv.SetDoneFunc(func(key tcell.Key) { app.SetRoot(BuildMainMenu(app), true) })
	app.SetRoot(tv, true)
}

func ShowBackupRestore(app *tview.Application) {
	form := tview.NewForm()

	var jsonText string
	var filePath string

	form.AddInputField("Backup JSON lines (optional)", "", 60, nil, func(text string) {
		jsonText = text
	}).
		AddInputField("Path to file (optional)", "", 60, nil, func(text string) {
			filePath = text
		}).
		AddButton(i18n.T("continue"), func() {
			var input string
			if filePath != "" {
				data, err := os.ReadFile(filePath)
				if err != nil {
					ShowMessageAndReturnToMenu(app, "Error reading file: %v", err)
					return
				}
				input = string(data)
			} else if jsonText != "" {
				input = jsonText
			} else {
				ShowMessageAndReturnToMenu(app, "Please provide JSON or file path.")
				return
			}

			lines := strings.Split(input, "\n")
			var updates []*v1.RelationshipUpdate
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				var r v1.Relationship
				if err := json.Unmarshal([]byte(line), &r); err == nil {
					updates = append(updates, &v1.RelationshipUpdate{
						Operation:    v1.RelationshipUpdate_OPERATION_CREATE,
						Relationship: &r,
					})
				} else {
					ShowMessageAndReturnToMenu(app, "Invalid JSON line: %v", err)
					return
				}
			}

			_, err := client.Client.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{Updates: updates})
			if err != nil {
				ShowMessageAndReturnToMenu(app, "Error restoring: %v", err)
			} else {
				ShowMessageAndReturnToMenu(app, "Backup restored successfully.")
			}
		}).
		AddButton(i18n.T("exit"), func() { app.SetRoot(BuildMainMenu(app), true) })

	form.SetBorder(true).SetTitle(i18n.T("backup_restore")).SetTitleAlign(tview.AlignLeft)
	AddFormReturnESC(form, app, func() { app.SetRoot(BuildMainMenu(app), true) })
	app.SetRoot(form, true)
}
