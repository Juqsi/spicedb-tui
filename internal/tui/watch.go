package tui

import (
	"context"
	"strings"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

const watchMaxLines = 1000

func ShowWatch(app *tview.Application) {
	showWatchForm(app, "")
}

func showWatchForm(app *tview.Application, oldTypes string) {
	form := tview.NewForm()
	form.
		AddInputField(i18n.T("watch_types"), oldTypes, 50, nil, nil).
		AddButton(i18n.T("start"), func() {
			typesText := form.GetFormItemByLabel(i18n.T("watch_types")).(*tview.InputField).GetText()
			var types []string
			for _, t := range strings.Split(typesText, ",") {
				if t = strings.TrimSpace(t); t != "" {
					types = append(types, t)
				}
			}
			startWatch(app, types)
		}).
		AddButton(i18n.T("exit"), func() { appPages.SwitchToPage("mainmenu") })
	form.SetBorder(true).SetTitle(i18n.T("watch_changes")).SetTitleAlign(tview.AlignLeft)
	AddEscBack(form, "mainmenu")
	appPages.AddAndSwitchToPage("watchform", form, true)
}

// watchSession holds the UI state of one watch page. All fields are only
// accessed from the UI goroutine (input handlers and QueueUpdateDraw callbacks).
type watchSession struct {
	app     *tview.Application
	types   []string
	ctx     context.Context
	cancel  context.CancelFunc
	view    *tview.TextView
	status  *tview.TextView
	token   *v1.ZedToken
	paused  bool
	pending []string
	running bool
	events  int
}

func startWatch(app *tview.Application, types []string) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &watchSession{app: app, types: types, ctx: ctx, cancel: cancel}

	w.view = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetMaxLines(watchMaxLines)
	title := " 📡 " + i18n.T("watch_changes") + " "
	w.view.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignLeft)
	w.status = tview.NewTextView().SetDynamicColors(true)

	w.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			w.cancel()
			appPages.SwitchToPage("mainmenu")
			return nil
		}
		if event.Key() != tcell.KeyRune {
			return event
		}
		switch event.Rune() {
		case 'p':
			w.paused = !w.paused
			if !w.paused {
				for _, l := range w.pending {
					w.write(l)
				}
				w.pending = nil
			}
			w.updateStatus()
			return nil
		case 'c':
			w.view.Clear()
			w.pending = nil
			return nil
		case 'r':
			if !w.running {
				w.run()
			}
			return nil
		}
		return event
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(w.view, 0, 1, true).
		AddItem(w.status, 1, 0, false)
	appPages.AddAndSwitchToPage("watch", layout, true)

	scope := i18n.T("all_types")
	if len(types) > 0 {
		scope = strings.Join(types, ", ")
	}
	w.write("[gray]" + tview.Escape(i18n.T("watch_started", scope)) + "[-]")
	w.run()
}

func (w *watchSession) write(line string) {
	_, _ = w.view.Write([]byte(line + "\n"))
}

// emit writes a line, or buffers it while paused.
func (w *watchSession) emit(line string) {
	w.events++
	if w.paused {
		w.pending = append(w.pending, line)
		if len(w.pending) > watchMaxLines {
			w.pending = w.pending[len(w.pending)-watchMaxLines:]
		}
		return
	}
	w.write(line)
}

func (w *watchSession) updateStatus() {
	s := "[green]● " + i18n.T("watch_live") + "[-]"
	switch {
	case !w.running:
		s = "[red]● " + i18n.T("watch_stopped") + "[-]"
	case w.paused:
		s = "[yellow]⏸ " + i18n.T("watch_paused", len(w.pending)) + "[-]"
	}
	s += "  " + i18n.T("watch_events", w.events)
	if w.token != nil {
		s += "  [gray]" + tview.Escape(i18n.T("watch_token", w.token.GetToken())) + "[-]"
	}
	s += "  [gray]" + i18n.T("watch_hint") + "[-]"
	w.status.SetText(s)
}

// run opens the watch stream (resuming from the last token, if any) and
// receives updates in a goroutine until the stream fails or the page is left.
func (w *watchSession) run() {
	w.running = true
	w.updateStatus()
	start := w.token
	ctx := w.ctx
	go func() {
		stream, err := client.Client.Watch(ctx, &v1.WatchRequest{
			OptionalObjectTypes: w.types,
			OptionalStartCursor: start,
		})
		for err == nil {
			var resp *v1.WatchResponse
			resp, err = stream.Recv()
			if err != nil {
				break
			}
			ts := time.Now().Format("15:04:05")
			var lines []string
			if resp.GetSchemaUpdated() {
				lines = append(lines, "[gray]"+ts+"[-]  [blue]"+i18n.T("watch_schema_updated")+"[-]")
			}
			for _, u := range resp.GetUpdates() {
				lines = append(lines, "[gray]"+ts+"[-]  "+watchOpLabel(u.GetOperation())+"  "+formatRelationship(u.GetRelationship()))
			}
			token := resp.GetChangesThrough()
			w.app.QueueUpdateDraw(func() {
				if token != nil {
					w.token = token
				}
				for _, l := range lines {
					w.emit(l)
				}
				w.updateStatus()
			})
		}
		if ctx.Err() != nil {
			return // page was left
		}
		w.app.QueueUpdateDraw(func() {
			w.running = false
			w.write("[red]" + tview.Escape(i18n.T("watch_error", err)) + "[-]")
			w.updateStatus()
		})
	}()
}

func watchOpLabel(op v1.RelationshipUpdate_Operation) string {
	switch op {
	case v1.RelationshipUpdate_OPERATION_CREATE:
		return "[green]CREATE[-]"
	case v1.RelationshipUpdate_OPERATION_TOUCH:
		return "[green]TOUCH [-]"
	case v1.RelationshipUpdate_OPERATION_DELETE:
		return "[red]DELETE[-]"
	}
	return "[gray]?     [-]"
}
