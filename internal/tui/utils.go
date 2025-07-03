package tui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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

func AsyncCall(app *tview.Application, loadingText string, fn func() string) {
	loadingView := tview.NewTextView().
		SetText(loadingText).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle("Loading...")

	app.SetRoot(loadingView, true)

	go func() {
		go func() {
			result := fn()
			app.QueueUpdateDraw(func() {
				var resultView *tview.TextView = tview.NewTextView().
					SetDynamicColors(true).
					SetText(result)
				resultView.SetDoneFunc(func(key tcell.Key) {
					app.SetRoot(BuildMainMenu(app), true)
				})
				resultView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					if event.Key() == tcell.KeyEsc {
						app.SetRoot(BuildMainMenu(app), true)
						return nil
					}
					return event
				})
				app.SetRoot(resultView, true)
			})
		}()
	}()
}
