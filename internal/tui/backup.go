package tui

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowBackupCreate() {
	tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	tv.SetBorder(true).SetTitle(i18n.T("backup_create"))

	schemaResp, err := client.Client.ReadSchema(context.Background(), &v1.ReadSchemaRequest{})
	if err != nil {
		tv.SetText(i18n.T("error_reading_schema", err))
		appPages.AddAndSwitchToPage("backupcreate", AddEscBack(tv, "mainmenu"), true)
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
			tv.Write([]byte(i18n.T("error_type", objType, err) + "\n"))
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
		tv.SetText(i18n.T("no_tuples_to_backup"))
	} else {
		tv.SetText(strings.Join(backupData, "\n"))
	}

	appPages.AddAndSwitchToPage("backupcreate", AddEscBack(tv, "mainmenu"), true)
}

func ShowBackupRestore() {
	form := tview.NewForm()
	var jsonText string
	var filePath string
	form.AddInputField(i18n.T("backup_json_input"), "", 60, nil, func(text string) {
		jsonText = text
	}).
		AddInputField(i18n.T("path_to_file"), "", 60, nil, func(text string) {
			filePath = text
		}).
		AddButton(i18n.T("continue"), func() {
			var input string
			if filePath != "" {
				data, err := os.ReadFile(filePath)
				if err != nil {
					ShowMessageAndReturnToMenu(i18n.T("error_reading_file", err))
					return
				}
				input = string(data)
			} else if jsonText != "" {
				input = jsonText
			} else {
				ShowMessageAndReturnToMenu(i18n.T("provide_json_or_file"))
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
					ShowMessageAndReturnToMenu(i18n.T("invalid_json_line", err))
					return
				}
			}

			_, err := client.Client.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{Updates: updates})
			if err != nil {
				ShowMessageAndReturnToMenu(i18n.T("error_restoring", err))
			} else {
				ShowMessageAndReturnToMenu(i18n.T("backup_restored_success"))
			}
		}).
		AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })

	form.SetBorder(true).SetTitle(i18n.T("backup_restore")).SetTitleAlign(tview.AlignLeft)
	AddEscBack(form, "mainmenu")
	appPages.AddAndSwitchToPage("backuprestore", form, true)
}
