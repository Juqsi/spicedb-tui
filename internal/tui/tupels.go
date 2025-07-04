package tui

import (
	"context"
	"fmt"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"strings"

	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func ShowAllTuples(app *tview.Application) {
	AsyncCallPages(app, i18n.T("loading"), func() (string, string) {
		var sb strings.Builder
		schemaResp, err := client.Client.ReadSchema(context.Background(), nil)
		if err != nil {
			return i18n.T("error_reading_schema", err), i18n.T("all_tuples")
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
			return i18n.T("no_object_types"), i18n.T("all_tuples")
		}
		total := 0
		for _, objectType := range objectTypes {
			sb.WriteString("\n[::b][white]📦 " + objectType + "[::-]\n")
			stream, err := client.Client.ReadRelationships(context.Background(), &v1.ReadRelationshipsRequest{
				RelationshipFilter: &v1.RelationshipFilter{ResourceType: objectType},
			})
			if err != nil {
				sb.WriteString("  [red]" + i18n.T("error_fetching", err) + "\n")
				continue
			}
			count := 0
			for {
				rel, err := stream.Recv()
				if err != nil {
					break
				}
				t := rel.GetRelationship()
				sb.WriteString(fmt.Sprintf("  %s:%s#%s@%s:%s\n",
					t.GetResource().GetObjectType(), t.GetResource().GetObjectId(),
					t.GetRelation(),
					t.GetSubject().GetObject().GetObjectType(), t.GetSubject().GetObject().GetObjectId(),
				))
				count++
				total++
			}
			if count == 0 {
				sb.WriteString("  [gray]" + i18n.T("no_tuples") + "\n")
			}
		}
		if total == 0 {
			sb.WriteString("\n[yellow]⚠️ " + i18n.T("no_tuples_in_db"))
		} else {
			sb.WriteString(fmt.Sprintf("\n[green]✔ " + i18n.T("loaded_tuples", total) + "\n"))
		}
		return sb.String(), i18n.T("all_tuples")
	})
}
