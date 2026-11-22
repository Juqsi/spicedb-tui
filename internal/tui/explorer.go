package tui

import (
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"spicedb-tui/internal/client"
	"spicedb-tui/internal/i18n"
)

const explorerPageSize = 500

// explorerFilter holds the raw filter inputs of the tuple explorer.
type explorerFilter struct {
	Resource string // "type" or "type:id"
	Relation string
	Subject  string // "type", "type:id" or "type:id#relation"
}

type explorerFocus int

const (
	explorerFocusTable explorerFocus = iota
	explorerFocusResource
	explorerFocusSubject
)

type tupleExplorer struct {
	app     *tview.Application
	onBack  func()
	filter  explorerFilter
	history []explorerFilter

	form     *tview.Form
	resField *tview.InputField
	relField *tview.InputField
	subField *tview.InputField
	table    *tview.Table
	status   *tview.TextView

	rows    []*v1.Relationship
	types   []string // resource types to iterate; nil until resolved
	typeIdx int
	cursor  *v1.Cursor
	more    bool
	loading bool
	errText string
	gen     int // incremented per query so stale results are dropped
}

// ShowTupleExplorer opens the explorer with focus on the filter form.
func ShowTupleExplorer(app *tview.Application) {
	openTupleExplorer(app, explorerFilter{}, explorerFocusResource, false, nil)
}

// ShowAllTuples opens the explorer and loads all tuples.
func ShowAllTuples(app *tview.Application) {
	openTupleExplorer(app, explorerFilter{}, explorerFocusTable, true, nil)
}

// ShowObjectRelations opens the explorer with focus on the resource filter.
func ShowObjectRelations(app *tview.Application) {
	openTupleExplorer(app, explorerFilter{}, explorerFocusResource, false, nil)
}

// ShowUserRelations opens the explorer with focus on the subject filter.
func ShowUserRelations(app *tview.Application) {
	openTupleExplorer(app, explorerFilter{}, explorerFocusSubject, false, nil)
}

// openTupleExplorer shows the explorer page. onBack is called on Esc (nil = main menu).
func openTupleExplorer(app *tview.Application, f explorerFilter, focus explorerFocus, run bool, onBack func()) {
	e := &tupleExplorer{app: app, onBack: onBack, filter: f}
	e.build()
	switch focus {
	case explorerFocusResource:
		e.form.SetFocus(0)
		app.SetFocus(e.form)
	case explorerFocusSubject:
		e.form.SetFocus(2)
		app.SetFocus(e.form)
	default:
		app.SetFocus(e.table)
	}
	if run {
		e.query()
	} else {
		e.render()
		e.updateStatus()
	}
}

func (e *tupleExplorer) build() {
	e.resField = tview.NewInputField().SetLabel(tview.Escape(i18n.T("filter_resource")) + " ").SetFieldWidth(22).SetText(e.filter.Resource)
	e.relField = tview.NewInputField().SetLabel(tview.Escape(i18n.T("filter_relation")) + " ").SetFieldWidth(14).SetText(e.filter.Relation)
	e.subField = tview.NewInputField().SetLabel(tview.Escape(i18n.T("filter_subject")) + " ").SetFieldWidth(22).SetText(e.filter.Subject)

	e.form = tview.NewForm().SetHorizontal(true).
		AddFormItem(e.resField).
		AddFormItem(e.relField).
		AddFormItem(e.subField).
		AddButton(i18n.T("apply"), e.apply)
	e.form.SetBorderPadding(0, 0, 1, 1)
	e.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			e.app.SetFocus(e.table)
			return nil
		}
		return event
	})

	e.table = tview.NewTable().SetSelectable(true, true).SetFixed(1, 0)
	e.table.SetBorder(true).SetTitleAlign(tview.AlignLeft)
	e.table.SetSelectedFunc(e.onSelect)
	e.table.SetInputCapture(e.tableKeys)

	e.status = tview.NewTextView().SetDynamicColors(true)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.form, 1, 0, false).
		AddItem(e.table, 0, 1, true).
		AddItem(e.status, 1, 0, false)
	appPages.AddAndSwitchToPage("explorer", layout, true)
}

// --- Navigation ---------------------------------------------------------

func (e *tupleExplorer) setFilter(f explorerFilter) {
	e.filter = f
	e.resField.SetText(f.Resource)
	e.relField.SetText(f.Relation)
	e.subField.SetText(f.Subject)
}

func (e *tupleExplorer) apply() {
	f := explorerFilter{
		Resource: strings.TrimSpace(e.resField.GetText()),
		Relation: strings.TrimSpace(e.relField.GetText()),
		Subject:  strings.TrimSpace(e.subField.GetText()),
	}
	if f != e.filter {
		e.history = append(e.history, e.filter)
	}
	e.filter = f
	e.query()
	e.app.SetFocus(e.table)
}

func (e *tupleExplorer) drill(f explorerFilter) {
	e.history = append(e.history, e.filter)
	e.setFilter(f)
	e.query()
}

func (e *tupleExplorer) back() {
	if len(e.history) == 0 {
		return
	}
	prev := e.history[len(e.history)-1]
	e.history = e.history[:len(e.history)-1]
	e.setFilter(prev)
	e.query()
}

func (e *tupleExplorer) exit() {
	if e.onBack != nil {
		e.onBack()
		return
	}
	appPages.SwitchToPage("mainmenu")
}

func (e *tupleExplorer) openTree(r *v1.Relationship) {
	openExpandTree(e.app, r.GetResource(), r.GetRelation(), func() {
		appPages.SwitchToPage("explorer")
	})
}

func (e *tupleExplorer) selectedRow() *v1.Relationship {
	row, _ := e.table.GetSelection()
	if row < 1 || row > len(e.rows) {
		return nil
	}
	return e.rows[row-1]
}

func (e *tupleExplorer) onSelect(row, col int) {
	r := e.selectedRow()
	if r == nil {
		return
	}
	switch col {
	case 0:
		e.drill(explorerFilter{Resource: formatObject(r.GetResource())})
	case 1:
		e.openTree(r)
	case 2:
		s := r.GetSubject()
		if rel := s.GetOptionalRelation(); rel != "" {
			// Subject set: show the tuples that define it, e.g. group:eng#member.
			e.drill(explorerFilter{Resource: formatObject(s.GetObject()), Relation: rel})
		} else {
			e.drill(explorerFilter{Subject: formatSubject(s)})
		}
	}
}

func (e *tupleExplorer) tableKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		e.exit()
		return nil
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.back()
		return nil
	case tcell.KeyTab:
		e.form.SetFocus(0)
		e.app.SetFocus(e.form)
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case '/':
			e.form.SetFocus(0)
			e.app.SetFocus(e.form)
			return nil
		case 'n':
			if e.more && !e.loading {
				e.fetch()
			}
			return nil
		case 'x':
			if r := e.selectedRow(); r != nil {
				e.openTree(r)
			}
			return nil
		}
	}
	return event
}

// --- Querying -----------------------------------------------------------

// query resets the result set and fetches the first page for the current filter.
func (e *tupleExplorer) query() {
	e.gen++
	e.rows = nil
	e.types = nil
	e.typeIdx = 0
	e.cursor = nil
	e.more = false
	e.errText = ""
	e.render()
	e.fetch()
}

func buildRelationshipFilter(f explorerFilter, resType string) *v1.RelationshipFilter {
	_, resID := parseTypeOptionalID(f.Resource)
	filter := &v1.RelationshipFilter{
		ResourceType:       resType,
		OptionalResourceId: resID,
		OptionalRelation:   f.Relation,
	}
	if st, sid, srel := parseSubjectFilter(f.Subject); st != "" {
		sf := &v1.SubjectFilter{SubjectType: st, OptionalSubjectId: sid}
		if srel != "" {
			sf.OptionalRelation = &v1.SubjectFilter_RelationFilter{Relation: srel}
		}
		filter.OptionalSubjectFilter = sf
	}
	return filter
}

// fetch loads the next page (up to explorerPageSize tuples). Without a resource
// type in the filter, it walks all definitions of the schema in order.
func (e *tupleExplorer) fetch() {
	gen := e.gen
	f := e.filter
	types, idx, cur := e.types, e.typeIdx, e.cursor
	e.loading = true
	e.updateStatus()

	go func() {
		ctx := context.Background()
		var got []*v1.Relationship
		var err error
		noTypes := false

		if types == nil {
			if resType, _ := parseTypeOptionalID(f.Resource); resType != "" {
				types = []string{resType}
			} else {
				types, err = schemaObjectTypes(ctx)
				noTypes = err == nil && len(types) == 0
			}
		}

		for err == nil && idx < len(types) && len(got) < explorerPageSize {
			want := explorerPageSize - len(got)
			stream, serr := client.Client.ReadRelationships(ctx, &v1.ReadRelationshipsRequest{
				Consistency:        fullyConsistent(),
				RelationshipFilter: buildRelationshipFilter(f, types[idx]),
				OptionalLimit:      uint32(want),
				OptionalCursor:     cur,
			})
			if serr != nil {
				err = serr
				break
			}
			n := 0
			for {
				resp, rerr := stream.Recv()
				if rerr == io.EOF {
					break
				}
				if rerr != nil {
					err = rerr
					break
				}
				got = append(got, resp.GetRelationship())
				cur = resp.GetAfterResultCursor()
				n++
			}
			if err != nil {
				break
			}
			if n < want { // this type is exhausted
				idx++
				cur = nil
			}
		}

		e.app.QueueUpdateDraw(func() {
			if gen != e.gen {
				return
			}
			first := len(e.rows) == 0
			e.loading = false
			e.types, e.typeIdx, e.cursor = types, idx, cur
			e.more = err == nil && idx < len(types)
			switch {
			case err != nil:
				e.errText = i18n.T("error_reading_relations", err)
			case noTypes:
				e.errText = i18n.T("no_object_types")
			}
			e.rows = append(e.rows, got...)
			e.render()
			if first && len(e.rows) > 0 {
				e.table.Select(1, 0).ScrollToBeginning()
			}
			e.updateStatus()
		})
	}()
}

// --- Rendering ----------------------------------------------------------

func (e *tupleExplorer) render() {
	e.table.Clear()
	headers := []string{
		i18n.T("col_resource"), i18n.T("col_relation"), i18n.T("col_subject"),
		i18n.T("col_caveat"), i18n.T("col_expires"),
	}
	for c, h := range headers {
		e.table.SetCell(0, c, tview.NewTableCell(tview.Escape(h)).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false))
	}
	for i, r := range e.rows {
		caveat, expires := "", ""
		if c := r.GetOptionalCaveat(); c != nil {
			caveat = c.GetCaveatName()
		}
		if x := r.GetOptionalExpiresAt(); x != nil {
			expires = x.AsTime().Format(time.RFC3339)
		}
		cells := []string{formatObject(r.GetResource()), r.GetRelation(), formatSubject(r.GetSubject()), caveat, expires}
		for c, txt := range cells {
			cell := tview.NewTableCell(tview.Escape(txt))
			switch c {
			case 1:
				cell.SetTextColor(tcell.ColorAqua)
			case 3:
				cell.SetTextColor(tcell.ColorYellow)
			case 4:
				cell.SetTextColor(tcell.ColorGray)
			}
			e.table.SetCell(i+1, c, cell)
		}
	}
	if len(e.rows) == 0 && !e.loading {
		e.table.SetCell(1, 0, tview.NewTableCell(tview.Escape(i18n.T("no_tuples"))).
			SetTextColor(tcell.ColorGray).SetSelectable(false))
	}
	e.table.SetTitle(" 🗂 " + tview.Escape(i18n.T("tuple_explorer")+"  "+e.filterString()) + " ")
}

func (e *tupleExplorer) filterString() string {
	var parts []string
	if e.filter.Resource != "" {
		parts = append(parts, "resource="+e.filter.Resource)
	}
	if e.filter.Relation != "" {
		parts = append(parts, "relation="+e.filter.Relation)
	}
	if e.filter.Subject != "" {
		parts = append(parts, "subject="+e.filter.Subject)
	}
	if len(parts) == 0 {
		return "*"
	}
	return strings.Join(parts, " ")
}

func (e *tupleExplorer) updateStatus() {
	s := i18n.T("explorer_count", len(e.rows))
	if e.more {
		s += "  [yellow]" + i18n.T("explorer_more") + "[-]"
	}
	if e.loading {
		s += "  ⏳ " + i18n.T("loading")
	}
	if e.errText != "" {
		s += "  [red]" + tview.Escape(e.errText) + "[-]"
	}
	if n := len(e.history); n > 0 {
		s += "  [gray]◂ " + strconv.Itoa(n) + "[-]"
	}
	s += "  [gray]" + i18n.T("explorer_hint") + "[-]"
	e.status.SetText(s)
}
