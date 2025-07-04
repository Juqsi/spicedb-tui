package tui

import (
	"context"
	"fmt"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"regexp"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
	"strings"
)

func ShowDeleteRelationsFiltered(app *tview.Application) {
	showDeleteRelationsFilteredForm(app, "", "", "", ".*")
}

func showDeleteRelationsFilteredForm(app *tview.Application, resVal, subVal, relVal, regexVal string) {
	form := tview.NewForm()
	appPages.AddAndSwitchToPage("preview", AddEscBack(form, "mainmenu"), true)
	form.
		AddInputField(i18n.T("resource_optional"), resVal, 30, nil, nil).
		AddInputField(i18n.T("subject_optional"), subVal, 30, nil, nil).
		AddInputField(i18n.T("relation_optional"), relVal, 20, nil, nil).
		AddInputField(i18n.T("id_regex"), regexVal, 30, nil, nil).
		AddButton(i18n.T("start_delete"), func() {
			res := form.GetFormItemByLabel(i18n.T("resource_optional")).(*tview.InputField).GetText()
			sub := form.GetFormItemByLabel(i18n.T("subject_optional")).(*tview.InputField).GetText()
			rel := form.GetFormItemByLabel(i18n.T("relation_optional")).(*tview.InputField).GetText()
			regexInput := form.GetFormItemByLabel(i18n.T("id_regex")).(*tview.InputField).GetText()

			loading := tview.NewTextView().
				SetText(i18n.T("searching_relations")).
				SetBorder(true).
				SetTitle(i18n.T("loading"))
			appPages.AddAndSwitchToPage("loading", loading, true)

			go func() {
				filter := &v1.RelationshipFilter{}
				if res != "" {
					r := strings.SplitN(res, ":", 2)
					filter.ResourceType = r[0]
					if len(r) == 2 {
						filter.OptionalResourceId = r[1]
					}
				}
				if sub != "" {
					s := strings.SplitN(sub, ":", 2)
					sf := &v1.SubjectFilter{
						SubjectType: s[0],
					}
					if len(s) == 2 {
						sf.OptionalSubjectId = s[1]
					}
					filter.OptionalSubjectFilter = sf
				}
				if rel != "" {
					filter.OptionalRelation = rel
				}

				re, err := regexp.Compile(regexInput)
				if err != nil {
					app.QueueUpdateDraw(func() {
						ShowMessageAndReturnToMenu(i18n.T("invalid_regex", err))
					})
					return
				}

				stream, err := client.Client.ReadRelationships(context.Background(), &v1.ReadRelationshipsRequest{
					RelationshipFilter: filter,
				})
				if err != nil {
					app.QueueUpdateDraw(func() {
						ShowMessageAndReturnToMenu(i18n.T("error_reading_relations", err))
					})
					return
				}

				var updates []*v1.RelationshipUpdate
				var previewLines []string

				for {
					rel, err := stream.Recv()
					if err != nil {
						break
					}
					r := rel.Relationship

					if res != "" && !re.MatchString(r.Resource.ObjectId) {
						continue
					}
					if sub != "" && !re.MatchString(r.Subject.Object.ObjectId) {
						continue
					}

					updates = append(updates, &v1.RelationshipUpdate{
						Operation:    v1.RelationshipUpdate_OPERATION_DELETE,
						Relationship: r,
					})

					if len(previewLines) < 100 {
						line := fmt.Sprintf("%s:%s#%s@%s:%s",
							r.Resource.ObjectType, r.Resource.ObjectId,
							r.Relation,
							r.Subject.Object.ObjectType, r.Subject.Object.ObjectId,
						)
						previewLines = append(previewLines, line)
					}
				}

				app.QueueUpdateDraw(func() {
					if len(updates) == 0 {
						ShowMessageAndReturnToMenu(i18n.T("no_relations_found"))
						return
					}

					text := i18n.T("will_delete_n_relations", len(updates)) + "\n\n"
					text += i18n.T("delete_confirm_hint") + "\n\n"
					if len(updates) > 100 {
						text += i18n.T("preview_first_n_only", 100) + "\n\n"
					}
					text += strings.Join(previewLines, "\n")

					resultView := tview.NewTextView().
						SetDynamicColors(true).
						SetScrollable(true).
						SetWrap(true).
						SetText(text)
					resultView.SetBorder(true).
						SetTitle(i18n.T("delete_preview_title"))

					resultView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
						switch event.Key() {
						case tcell.KeyCtrlD:
							appPages.AddAndSwitchToPage("loading", tview.NewTextView().SetText(i18n.T("deleting_relation")).SetBorder(true).SetTitle(i18n.T("deleting_relation_title")), true)
							go func() {
								_, err := client.Client.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
									Updates: updates,
								})
								app.QueueUpdateDraw(func() {
									if err != nil {
										ShowMessageAndReturnToMenu(i18n.T("error_deleting_relation", err))
									} else {
										ShowMessageAndReturnToMenu(i18n.T("relation_deleted_success_n", len(updates)))
									}
								})
							}()
							return nil
						case tcell.KeyEsc:
							showDeleteRelationsFilteredForm(app, res, sub, rel, regexInput)
							return nil
						}
						return event
					})
					appPages.AddAndSwitchToPage("preview", resultView, true)
				})
			}()
		}).
		AddButton(i18n.T("back"), func() {
			appPages.SwitchToPage("mainmenu")
		})

	form.SetBorder(true).
		SetTitle(i18n.T("delete_relation_filtered_title")).
		SetTitleAlign(tview.AlignLeft)

	appPages.AddAndSwitchToPage("form", form, true)
}
