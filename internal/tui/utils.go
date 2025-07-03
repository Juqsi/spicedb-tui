package tui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/i18n"
)

func ShowMessageAndReturnToMenu(app *tview.Application, msg string, args ...interface{}) {
	text := fmt.Sprintf(msg, args...)
	tv := tview.NewTextView().SetText(text).
		SetDoneFunc(func(key tcell.Key) {
			app.SetRoot(BuildMainMenu(app), true)
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				app.SetRoot(BuildMainMenu(app), true)
				return nil
			}
			return event
		})
	app.SetRoot(tv, true)
}

func AddFormReturnESC(form *tview.Form, app *tview.Application, menuFunc func()) {
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			menuFunc()
			return nil
		}
		return event
	})
}

func AsyncCall(app *tview.Application, loadingText string, fn func() (result string, title string)) {
	loadingView := tview.NewTextView().
		SetText(loadingText).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle("⏳ " + i18n.T("loading"))

	app.SetRoot(loadingView, true)

	go func() {
		resultText, resultTitle := fn()

		app.QueueUpdateDraw(func() {
			tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
			tv.SetBorder(true).
				SetTitle(resultTitle)
			tv.SetText(resultText)
			tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					app.SetRoot(BuildMainMenu(app), true)
					return nil
				}
				return event
			})

			app.SetRoot(tv, true)
		})
	}()
}
