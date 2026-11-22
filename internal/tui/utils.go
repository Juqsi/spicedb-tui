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

// showLoadingPage switches to a simple bordered "loading" page.
func showLoadingPage(text string) {
	loadingView := tview.NewTextView().
		SetText(text).
		SetDynamicColors(true)
	loadingView.SetBorder(true).SetTitle("⏳ " + i18n.T("loading"))
	appPages.AddAndSwitchToPage("loading", loadingView, true)
}

// showTextResult shows a scrollable text page; Esc calls onBack.
func showTextResult(pageName, title, text string, onBack func()) {
	tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	tv.SetBorder(true).SetTitle(title)
	tv.SetText(text)
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			onBack()
			return nil
		}
		return event
	})
	appPages.AddAndSwitchToPage(pageName, tv, true)
}

func AsyncCallPages(app *tview.Application, loadingText string, fn func() (result string, title string)) {
	showLoadingPage(loadingText)
	go func() {
		resultText, resultTitle := fn()
		app.QueueUpdateDraw(func() {
			showTextResult("asyncresult", resultTitle, resultText, func() { appPages.SwitchToPage("mainmenu") })
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

// fullyConsistent requests evaluation at the newest revision, so introspection
// screens show writes immediately instead of a cached (quantized) snapshot.
func fullyConsistent() *v1.Consistency {
	return &v1.Consistency{Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true}}
}

// --- Formatting ---------------------------------------------------------

// formatObject renders "type:id" (raw, not escaped).
func formatObject(o *v1.ObjectReference) string {
	return o.GetObjectType() + ":" + o.GetObjectId()
}

// formatSubject renders "type:id" plus "#relation" for subject sets (raw, not escaped).
func formatSubject(s *v1.SubjectReference) string {
	out := formatObject(s.GetObject())
	if rel := s.GetOptionalRelation(); rel != "" {
		out += "#" + rel
	}
	return out
}

// formatRelationship renders a tuple for display in views with dynamic colors.
// The tuple text is escaped; caveat and expiration are appended as colored hints.
func formatRelationship(r *v1.Relationship) string {
	out := tview.Escape(formatObject(r.GetResource()) + "#" + r.GetRelation() + "@" + formatSubject(r.GetSubject()))
	if c := r.GetOptionalCaveat(); c != nil {
		out += " [yellow](caveat: " + tview.Escape(c.GetCaveatName()) + ")[-]"
	}
	if e := r.GetOptionalExpiresAt(); e != nil {
		out += " [gray](expires " + e.AsTime().Format(time.RFC3339) + ")[-]"
	}
	return out
}

// --- Parsing ------------------------------------------------------------

// parseObjectRef parses "type:id"; both parts must be non-empty.
func parseObjectRef(s string) (*v1.ObjectReference, bool) {
	p := strings.SplitN(strings.TrimSpace(s), ":", 2)
	if len(p) != 2 || p[0] == "" || p[1] == "" {
		return nil, false
	}
	return &v1.ObjectReference{ObjectType: p[0], ObjectId: p[1]}, true
}

// parseSubjectRef parses "type:id" or "type:id#relation".
func parseSubjectRef(s string) (*v1.SubjectReference, bool) {
	obj, rel, _ := strings.Cut(strings.TrimSpace(s), "#")
	ref, ok := parseObjectRef(obj)
	if !ok {
		return nil, false
	}
	return &v1.SubjectReference{Object: ref, OptionalRelation: rel}, true
}

// parseTypeOptionalID parses "type" or "type:id".
func parseTypeOptionalID(s string) (typ, id string) {
	typ, id, _ = strings.Cut(strings.TrimSpace(s), ":")
	return typ, id
}

// parseSubjectFilter parses "type", "type:id" or "type:id#relation".
func parseSubjectFilter(s string) (typ, id, rel string) {
	obj, rel, _ := strings.Cut(strings.TrimSpace(s), "#")
	typ, id = parseTypeOptionalID(obj)
	return typ, id, rel
}

// schemaObjectTypes returns the names of all definitions in the current schema.
func schemaObjectTypes(ctx context.Context) ([]string, error) {
	rsp, err := client.Client.ReadSchema(ctx, &v1.ReadSchemaRequest{})
	if err != nil {
		return nil, err
	}
	var types []string
	for _, line := range strings.Split(rsp.SchemaText, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "definition ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				types = append(types, strings.TrimSuffix(parts[1], "{"))
			}
		}
	}
	return types, nil
}
