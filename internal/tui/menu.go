package tui

import (
	"github.com/rivo/tview"
	"os"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

func StartTUI(app *tview.Application) {
	if err := client.InitClient(); err != nil {
		ShowMessageAndReturnToMenu(app, "Connection failed: %v", err)
		return
	}
	menu := BuildMainMenu(app)
	app.SetRoot(menu, true)
}

func BuildMainMenu(app *tview.Application) *tview.List {
	menu := tview.NewList().
		AddItem(i18n.T("show_schema"), "", 's', func() { ShowSchema(app) }).
		AddItem(i18n.T("upload_schema"), "", 'w', func() { ShowWriteSchema(app) }).
		AddItem(i18n.T("object_relations"), "", 'o', func() { ShowObjectRelations(app) }).
		AddItem(i18n.T("user_relations"), "", 'u', func() { ShowUserRelations(app) }).
		AddItem(i18n.T("all_tuples"), "", 'a', func() { ShowAllTuples(app) }).
		AddItem(i18n.T("add_relation"), "", 'r', func() { ShowAddRelation(app) }).
		AddItem(i18n.T("permission_check"), "", 'p', func() { ShowPermissionCheck(app) }).
		AddItem(i18n.T("backup_create"), "", 'b', func() { ShowBackupCreate(app) }).
		AddItem(i18n.T("backup_restore"), "", 'l', func() { ShowBackupRestore(app) }).
		AddItem(i18n.T("data_import"), "", 'i', func() { ShowDataImport(app) }).
		AddItem(i18n.T("quit"), "", 'q', func() { confirmExit(app) })

	menu.SetBorder(true).SetTitle(i18n.T("app_title")).SetTitleAlign(tview.AlignLeft)
	return menu
}

func confirmExit(app *tview.Application) {
	modal := tview.NewModal().
		SetText(i18n.T("confirm_exit")).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				app.Stop()
				os.Exit(0)
			} else {
				app.SetRoot(BuildMainMenu(app), true)
			}
		})
	app.SetRoot(modal, true)
}
