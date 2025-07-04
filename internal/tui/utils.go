package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/i18n"
)

func ShowMessageAndReturnToMenu(msg string, args ...interface{}) {
	text := i18n.T(msg, args...)
	tv := tview.NewTextView().
		SetText(text).
		SetDynamicColors(true).
		SetDoneFunc(func(key tcell.Key) {
			appPages.SwitchToPage("mainmenu")
		})
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyRune {
			appPages.SwitchToPage("mainmenu")
			return nil
		}
		return event
	})
	tv.SetBorder(true).SetTitle(i18n.T("message"))
	appPages.AddAndSwitchToPage("msgpage", tv, true)
}

func AsyncCallPages(app *tview.Application, loadingText string, fn func() (result string, title string)) {
	loadingView := tview.NewTextView().
		SetText(loadingText).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle("⏳ " + i18n.T("loading"))
	appPages.AddAndSwitchToPage("loading", loadingView, true)
	go func() {
		resultText, resultTitle := fn()
		app.QueueUpdateDraw(func() {
			tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
			tv.SetBorder(true).
				SetTitle(resultTitle)
			tv.SetText(resultText)
			tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					appPages.SwitchToPage("mainmenu")
					return nil
				}
				return event
			})
			appPages.AddAndSwitchToPage("asyncresult", tv, true)
		})
	}()
}

func AddEscBack(prim tview.Primitive, toPage string) tview.Primitive {
	if tv, ok := prim.(*tview.TextView); ok {
		oldIC := tv.GetInputCapture()
		tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				appPages.SwitchToPage(toPage)
				return nil
			}
			if oldIC != nil {
				return oldIC(event)
			}
			return event
		})
	}
	if f, ok := prim.(*tview.Form); ok {
		oldIC := f.GetInputCapture()
		f.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				appPages.SwitchToPage(toPage)
				return nil
			}
			if oldIC != nil {
				return oldIC(event)
			}
			return event
		})
	}
	return prim
}
