package planner

// Test cases against project.md. Each test names the requirement it covers.

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// newTestService returns a service persisting to a temp file.
func newTestService(t *testing.T) (*Service, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.json")
	s, err := NewService(path)
	if err != nil {
		t.Fatal(err)
	}
	return s, path
}

func mustCategory(t *testing.T, s *Service, kind Kind, name string) Category {
	t.Helper()
	if _, err := s.AddCategory(kind, name); err != nil {
		t.Fatalf("AddCategory(%s): %v", name, err)
	}
	c := s.findCategoryByName(kind, name)
	if c == nil {
		t.Fatalf("category %s not found after adding", name)
	}
	return *c
}

func mustEntry(t *testing.T, s *Service, cat Category, name string, cents int64, period Period) Entry {
	t.Helper()
	_, err := s.AddEntry(EntryInput{CategoryID: cat.ID, Name: name, AmountCents: cents, Period: period})
	if err != nil {
		t.Fatalf("AddEntry(%s): %v", name, err)
	}
	for _, e := range s.data.EntriesOf(cat.ID) {
		if e.Name == name {
			return e
		}
	}
	t.Fatalf("entry %s not found after adding", name)
	return Entry{}
}

func names(entries []EntryView) []string {
	out := []string{}
	for _, e := range entries {
		out = append(out, e.Name)
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// "Categories have to be added first, before spendings can be entered."
func TestEntryRequiresExistingCategory(t *testing.T) {
	s, _ := newTestService(t)
	if _, err := s.AddEntry(EntryInput{CategoryID: "does-not-exist", Name: "Rent", AmountCents: 100000, Period: PeriodMonthly}); err == nil {
		t.Fatal("expected error when adding an entry without a category")
	}
	cat := mustCategory(t, s, KindSpending, "Housing")
	if _, err := s.AddEntry(EntryInput{CategoryID: cat.ID, Name: "Rent", AmountCents: 100000, Period: PeriodMonthly}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// "Each spending has a name, an amount in € and a category."
func TestEntryValidation(t *testing.T) {
	s, _ := newTestService(t)
	cat := mustCategory(t, s, KindSpending, "Housing")
	cases := []EntryInput{
		{Name: "", AmountCents: 100, Period: PeriodMonthly},
		{Name: "   ", AmountCents: 100, Period: PeriodMonthly},
		{Name: "Rent", AmountCents: 0, Period: PeriodMonthly},
		{Name: "Rent", AmountCents: -5, Period: PeriodMonthly},
		{Name: "Rent", AmountCents: 100, Period: Period("weekly")},
		{Name: "Rent", AmountCents: 100, Period: PeriodYearly, DueMonth: 13},
		{Name: "Rent", AmountCents: 100, Period: PeriodYearly, DueMonth: -1},
	}
	for _, c := range cases {
		c.CategoryID = cat.ID
		if _, err := s.AddEntry(c); err == nil {
			t.Errorf("expected validation error for %+v", c)
		}
	}
	if _, err := s.AddCategory(KindSpending, "  "); err == nil {
		t.Error("expected error for empty category name")
	}
	if _, err := s.AddCategory(Kind("other"), "X"); err == nil {
		t.Error("expected error for unknown kind")
	}
	if _, err := s.AddCategory(KindSpending, "housing"); err == nil {
		t.Error("expected error for duplicate category name (case-insensitive)")
	}
}

// "A Category can't be deleted unless it is not used anymore."
func TestCategoryDeleteOnlyWhenUnused(t *testing.T) {
	s, _ := newTestService(t)
	cat := mustCategory(t, s, KindSpending, "Housing")
	e := mustEntry(t, s, cat, "Rent", 100000, PeriodMonthly)

	if _, err := s.DeleteCategory(cat.ID); err == nil {
		t.Fatal("expected error deleting a used category")
	}
	if s.data.Category(cat.ID) == nil {
		t.Fatal("category must still exist after failed delete")
	}

	// "A spending can be deleted."
	if _, err := s.DeleteEntry(e.ID); err != nil {
		t.Fatal(err)
	}
	st, err := s.DeleteCategory(cat.ID)
	if err != nil {
		t.Fatalf("deleting unused category: %v", err)
	}
	if len(st.Spending) != 0 {
		t.Fatalf("expected no spending categories, got %d", len(st.Spending))
	}
}

// "Per entry it shows how much it is per month and yearly ... the missing
// value is calculated."
func TestMonthlyYearlyConversion(t *testing.T) {
	monthly := Entry{AmountCents: 12345, Period: PeriodMonthly}
	if got := monthly.YearlyCents(); got != 12345*12 {
		t.Errorf("monthly->yearly = %d, want %d", got, 12345*12)
	}
	if got := monthly.MonthlyCents(); got != 12345 {
		t.Errorf("monthly stays %d, got %d", 12345, got)
	}
	yearly := Entry{AmountCents: 120000, Period: PeriodYearly}
	if got := yearly.MonthlyCents(); got != 10000 {
		t.Errorf("yearly->monthly = %d, want %d", got, 10000)
	}
	if got := yearly.YearlyCents(); got != 120000 {
		t.Errorf("yearly stays %d, got %d", 120000, got)
	}
	// 100,00 € / 12 = 8,333.. € -> rounded to 8,33 €
	odd := Entry{AmountCents: 10000, Period: PeriodYearly}
	if got := odd.MonthlyCents(); got != 833 {
		t.Errorf("10000/12 rounded = %d, want 833", got)
	}
	// 50,00 € / 12 = 4,1666.. € -> rounded to 4,17 €
	odd2 := Entry{AmountCents: 5000, Period: PeriodYearly}
	if got := odd2.MonthlyCents(); got != 417 {
		t.Errorf("5000/12 rounded = %d, want 417", got)
	}
}

// "It should be visible if it is a monthly or yearly entry" - the period is
// the master and must survive the round trip through the state and the file.
func TestPeriodIsPersistedAsMaster(t *testing.T) {
	s, path := newTestService(t)
	cat := mustCategory(t, s, KindSpending, "Insurance")
	mustEntry(t, s, cat, "Car", 60000, PeriodYearly)
	mustEntry(t, s, cat, "Health", 20000, PeriodMonthly)

	st := s.GetState()
	if st.Spending[0].Entries[0].Period != PeriodYearly || st.Spending[0].Entries[1].Period != PeriodMonthly {
		t.Fatalf("periods not preserved in state: %+v", st.Spending[0].Entries)
	}
	if st.Spending[0].Entries[0].MonthlyCents != 5000 || st.Spending[0].Entries[0].YearlyCents != 60000 {
		t.Fatalf("yearly entry view wrong: %+v", st.Spending[0].Entries[0])
	}
	if st.Spending[0].MonthlyCents != 25000 || st.Spending[0].YearlyCents != 300000 {
		t.Fatalf("category subtotal wrong: %d / %d", st.Spending[0].MonthlyCents, st.Spending[0].YearlyCents)
	}

	reloaded, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Entries[0].Period != PeriodYearly {
		t.Fatal("period lost after reload")
	}
}

// "The main application view ... groups them by category ... income first,
// spendings below."
func TestStateGroupsByKindAndCategory(t *testing.T) {
	s, _ := newTestService(t)
	salary := mustCategory(t, s, KindIncome, "Salary")
	housing := mustCategory(t, s, KindSpending, "Housing")
	mustEntry(t, s, salary, "Job", 300000, PeriodMonthly)
	mustEntry(t, s, housing, "Rent", 100000, PeriodMonthly)
	mustEntry(t, s, housing, "Electricity", 8000, PeriodMonthly)

	st := s.GetState()
	if len(st.Income) != 1 || st.Income[0].Name != "Salary" || len(st.Income[0].Entries) != 1 {
		t.Fatalf("income block wrong: %+v", st.Income)
	}
	if len(st.Spending) != 1 || len(st.Spending[0].Entries) != 2 {
		t.Fatalf("spending block wrong: %+v", st.Spending)
	}
}

// Statistics box: average cost per month, saldo per month, saldo per year,
// transfer to bank account (monthly spendings) and to savings account (1/12
// of yearly spendings).
func TestStatistics(t *testing.T) {
	s, _ := newTestService(t)
	salary := mustCategory(t, s, KindIncome, "Salary")
	housing := mustCategory(t, s, KindSpending, "Housing")
	insurance := mustCategory(t, s, KindSpending, "Insurance")
	mustEntry(t, s, salary, "Job", 300000, PeriodMonthly)      // 3000/month
	mustEntry(t, s, salary, "Bonus", 120000, PeriodYearly)     // 1200/year = 100/month
	mustEntry(t, s, housing, "Rent", 100000, PeriodMonthly)    // 1000/month
	mustEntry(t, s, insurance, "Car", 60000, PeriodYearly)     // 600/year = 50/month
	mustEntry(t, s, insurance, "Health", 20000, PeriodMonthly) // 200/month

	st := s.GetState().Stats
	if len(st.TopSpendings) != 3 || st.TopSpendings[0].Name != "Rent" {
		t.Fatalf("top spendings: %+v", st.TopSpendings)
	}
	st.TopSpendings = nil // checked above; the struct compare needs comparable fields
	want := Stats{
		IncomeMonthlyCents:      310000,
		IncomeYearlyCents:       3720000,
		SpendingMonthlyCents:    125000,
		SpendingYearlyCents:     1500000,
		SaldoMonthlyCents:       185000,
		SaldoYearlyCents:        2220000,
		ToBankMonthlyCents:      120000,
		ToSavingsMonthlyCents:   5000,
		RemainingAfterGoalCents: 185000, // no goal set: equals the saldo
		GoalReachable:           true,
		UnscheduledCount:        1, // "Car" has no due month
	}
	if !reflect.DeepEqual(st, want) {
		t.Fatalf("stats\n got %+v\nwant %+v", st, want)
	}
	if st.ToBankMonthlyCents+st.ToSavingsMonthlyCents != st.SpendingMonthlyCents {
		t.Fatal("bank + savings must equal the average cost per month")
	}
}

// "Income and Spending categories should be sortable via drag & drop. This is
// saved as well."
func TestMoveCategory(t *testing.T) {
	s, path := newTestService(t)
	a := mustCategory(t, s, KindSpending, "A")
	mustCategory(t, s, KindIncome, "Salary")
	mustCategory(t, s, KindSpending, "B")
	c := mustCategory(t, s, KindSpending, "C")

	catNames := func(views []CategoryView) []string {
		out := []string{}
		for _, v := range views {
			out = append(out, v.Name)
		}
		return out
	}

	st, err := s.MoveCategory(c.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := catNames(st.Spending); !equalStrings(got, []string{"C", "A", "B"}) {
		t.Fatalf("after moving C to front: %v", got)
	}
	st, _ = s.MoveCategory(a.ID, 5) // past the end -> last
	if got := catNames(st.Spending); !equalStrings(got, []string{"C", "B", "A"}) {
		t.Fatalf("after moving A to end: %v", got)
	}
	st, _ = s.MoveCategory(a.ID, 1)
	if got := catNames(st.Spending); !equalStrings(got, []string{"C", "A", "B"}) {
		t.Fatalf("after moving A to middle: %v", got)
	}
	// Income block is untouched by spending moves.
	if got := catNames(st.Income); !equalStrings(got, []string{"Salary"}) {
		t.Fatalf("income block changed: %v", got)
	}
	// Order is persisted.
	reloaded, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := catNames(reloaded.BuildViews(KindSpending)); !equalStrings(got, []string{"C", "A", "B"}) {
		t.Fatalf("order not persisted: %v", got)
	}
}

// "Entries themselves should be sortable as well via drag & drop - within the
// respective category." and "drag & drop an entry to a different category".
func TestMoveEntry(t *testing.T) {
	s, path := newTestService(t)
	housing := mustCategory(t, s, KindSpending, "Housing")
	leisure := mustCategory(t, s, KindSpending, "Leisure")
	salary := mustCategory(t, s, KindIncome, "Salary")
	rent := mustEntry(t, s, housing, "Rent", 100000, PeriodMonthly)
	mustEntry(t, s, housing, "Power", 8000, PeriodMonthly)
	water := mustEntry(t, s, housing, "Water", 3000, PeriodMonthly)
	mustEntry(t, s, leisure, "Gym", 4000, PeriodMonthly)

	// Reorder within the category.
	st, err := s.MoveEntry(water.ID, housing.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := names(st.Spending[0].Entries); !equalStrings(got, []string{"Water", "Rent", "Power"}) {
		t.Fatalf("after reorder: %v", got)
	}
	// Move to a different category, at the front.
	st, err = s.MoveEntry(rent.ID, leisure.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := names(st.Spending[0].Entries); !equalStrings(got, []string{"Water", "Power"}) {
		t.Fatalf("source after move: %v", got)
	}
	if got := names(st.Spending[1].Entries); !equalStrings(got, []string{"Rent", "Gym"}) {
		t.Fatalf("target after move: %v", got)
	}
	// Move to the end of a different category.
	st, _ = s.MoveEntry(water.ID, leisure.ID, 99)
	if got := names(st.Spending[1].Entries); !equalStrings(got, []string{"Rent", "Gym", "Water"}) {
		t.Fatalf("target after move to end: %v", got)
	}
	// Entries can't cross the income/spending border.
	if _, err := s.MoveEntry(water.ID, salary.ID, 0); err == nil {
		t.Fatal("expected error moving a spending into an income category")
	}
	// Persisted.
	reloaded, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := names(reloaded.BuildViews(KindSpending)[1].Entries); !equalStrings(got, []string{"Rent", "Gym", "Water"}) {
		t.Fatalf("entry order not persisted: %v", got)
	}
}

// Editing an entry, including changing its category.
func TestUpdateEntry(t *testing.T) {
	s, _ := newTestService(t)
	housing := mustCategory(t, s, KindSpending, "Housing")
	leisure := mustCategory(t, s, KindSpending, "Leisure")
	rent := mustEntry(t, s, housing, "Rent", 100000, PeriodMonthly)

	st, err := s.UpdateEntry(rent.ID, EntryInput{
		CategoryID: leisure.ID, Name: "Rent (new)", AmountCents: 1200000, Period: PeriodYearly,
		DueMonth: 3, Notes: " contract 42 ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Spending[0].Entries) != 0 || len(st.Spending[1].Entries) != 1 {
		t.Fatalf("entry not moved: %+v", st.Spending)
	}
	e := st.Spending[1].Entries[0]
	if e.Name != "Rent (new)" || e.AmountCents != 1200000 || e.Period != PeriodYearly || e.MonthlyCents != 100000 {
		t.Fatalf("entry not updated: %+v", e)
	}
	if e.DueMonth != 3 || e.Notes != "contract 42" {
		t.Fatalf("due month / notes not stored: %+v", e)
	}
	// Switching back to monthly clears the due month.
	st, _ = s.UpdateEntry(rent.ID, EntryInput{CategoryID: leisure.ID, Name: "Rent", AmountCents: 100, Period: PeriodMonthly, DueMonth: 3})
	if st.Spending[1].Entries[0].DueMonth != 0 {
		t.Fatal("monthly entries must not keep a due month")
	}
	if _, err := s.UpdateEntry(rent.ID, EntryInput{CategoryID: leisure.ID, Name: "", AmountCents: 1, Period: PeriodMonthly}); err == nil {
		t.Fatal("expected validation error")
	}
}

// "Categories should be collapsible, and at the top of both blocks (i.e.
// income and spending) there should be a 'expand all', 'collapse all'
// function."
func TestCollapse(t *testing.T) {
	s, path := newTestService(t)
	a := mustCategory(t, s, KindSpending, "A")
	mustCategory(t, s, KindSpending, "A2")
	mustCategory(t, s, KindIncome, "B")

	st, err := s.SetCategoryCollapsed(a.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Spending[0].Collapsed || st.Spending[1].Collapsed || st.Income[0].Collapsed {
		t.Fatal("single collapse wrong")
	}
	// Collapse all only affects the block it was triggered in.
	st, err = s.SetAllCollapsed(KindSpending, true)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Spending[0].Collapsed || !st.Spending[1].Collapsed || st.Income[0].Collapsed {
		t.Fatal("collapse all spending wrong")
	}
	st, _ = s.SetAllCollapsed(KindIncome, true)
	if !st.Income[0].Collapsed {
		t.Fatal("collapse all income wrong")
	}
	st, _ = s.SetAllCollapsed(KindSpending, false)
	if st.Spending[0].Collapsed || st.Spending[1].Collapsed || !st.Income[0].Collapsed {
		t.Fatal("expand all spending wrong")
	}
	if _, err := s.SetAllCollapsed(Kind("other"), true); err == nil {
		t.Fatal("expected error for unknown kind")
	}
	reloaded, _ := LoadFile(path)
	if reloaded.Categories[0].Collapsed || !reloaded.Categories[2].Collapsed {
		t.Fatal("collapsed state not persisted")
	}
}

// "persist them to disk as JSON ... a single structured JSON file ... The JSON
// file should carry a version number"
func TestPersistenceAndVersion(t *testing.T) {
	s, path := newTestService(t)
	cat := mustCategory(t, s, KindIncome, "Salary")
	mustEntry(t, s, cat, "Job", 300000, PeriodMonthly)

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file map[string]any
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("data.json is not valid JSON: %v", err)
	}
	if v, _ := file["version"].(float64); int(v) != CurrentVersion {
		t.Fatalf("version = %v, want %d", file["version"], CurrentVersion)
	}
	if _, ok := file["categories"]; !ok {
		t.Fatal("categories missing")
	}
	if _, ok := file["entries"]; !ok {
		t.Fatal("entries missing")
	}

	// A second service on the same file sees the same data.
	s2, err := NewService(path)
	if err != nil {
		t.Fatal(err)
	}
	st := s2.GetState()
	if len(st.Income) != 1 || st.Income[0].Entries[0].Name != "Job" {
		t.Fatalf("data not reloaded: %+v", st)
	}
	// No temp files are left behind by the atomic write.
	matches, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".data-*"))
	if len(matches) != 0 {
		t.Fatalf("temp files left behind: %v", matches)
	}
}

// "The JSON file should be stored next to the binary and be called data.json"
// - the file exists as soon as the service starts.
func TestDataFileIsCreatedOnStartup(t *testing.T) {
	_, path := newTestService(t)
	if filepath.Base(path) != DefaultDataFileName {
		t.Fatalf("file name = %s", filepath.Base(path))
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("data.json not created on startup: %v", err)
	}
	if _, err := Decode(raw); err != nil {
		t.Fatalf("created file invalid: %v", err)
	}
}

// Merge matches categories by id before name and never duplicates entries
// whose id already exists.
func TestMergeMatchesByID(t *testing.T) {
	base := NewData()
	base.Categories = []Category{
		{ID: "cat-housing", Name: "Housing", Kind: KindSpending},
		{ID: "cat-salary", Name: "Salary", Kind: KindIncome},
	}
	base.Entries = []Entry{
		{ID: "e-rent", CategoryID: "cat-housing", Name: "Rent", AmountCents: 100000, Period: PeriodMonthly},
	}

	extra := NewData()
	extra.Categories = []Category{
		// Same id, renamed in the file -> reused by id, keeps the base name.
		{ID: "cat-housing", Name: "Home", Kind: KindSpending},
		// Different id, same name -> reused by name.
		{ID: "cat-other", Name: "salary", Kind: KindIncome},
		// Unknown id and name -> added, keeps its id.
		{ID: "cat-leisure", Name: "Leisure", Kind: KindSpending},
		// Same id as an existing category but a different kind -> not the
		// same category; added under a fresh id.
		{ID: "cat-salary", Name: "Bonus", Kind: KindSpending},
	}
	extra.Entries = []Entry{
		{ID: "e-rent", CategoryID: "cat-housing", Name: "Rent (changed)", AmountCents: 1, Period: PeriodYearly}, // skipped
		{ID: "e-power", CategoryID: "cat-housing", Name: "Power", AmountCents: 8000, Period: PeriodMonthly},     // added to Housing
		{ID: "e-job", CategoryID: "cat-other", Name: "Job", AmountCents: 300000, Period: PeriodMonthly},         // added to Salary
		{ID: "e-gym", CategoryID: "cat-leisure", Name: "Gym", AmountCents: 4000, Period: PeriodMonthly},         // added to Leisure
		{ID: "e-bonus", CategoryID: "cat-salary", Name: "Bonus", AmountCents: 50000, Period: PeriodYearly},      // added to spending "Bonus"
	}

	merged, report := Merge(base, extra)
	if err := merged.Validate(); err != nil {
		t.Fatalf("merged data invalid: %v", err)
	}
	if report.CategoriesAdded != 2 || report.CategoriesReused != 2 || report.EntriesAdded != 4 || report.EntriesSkipped != 1 {
		t.Fatalf("report wrong: %+v", report)
	}

	housing := merged.Category("cat-housing")
	if housing == nil || housing.Name != "Housing" {
		t.Fatalf("housing category not reused by id: %+v", housing)
	}
	if got := names(merged.BuildViews(KindSpending)[0].Entries); !equalStrings(got, []string{"Rent", "Power"}) {
		t.Fatalf("housing entries: %v", got)
	}
	if rent := merged.Entry("e-rent"); rent.Name != "Rent" || rent.AmountCents != 100000 {
		t.Fatalf("existing entry was modified: %+v", rent)
	}
	if got := names(merged.BuildViews(KindIncome)[0].Entries); !equalStrings(got, []string{"Job"}) || len(merged.CategoriesOf(KindIncome)) != 1 {
		t.Fatalf("salary not reused by name: %+v", merged.BuildViews(KindIncome))
	}
	if merged.Category("cat-leisure") == nil {
		t.Fatal("new category should keep its imported id")
	}
	bonus := merged.Entry("e-bonus")
	if bonus == nil || bonus.CategoryID == "cat-salary" || merged.Category(bonus.CategoryID).Kind != KindSpending {
		t.Fatalf("kind-mismatched category id must not be reused: %+v", bonus)
	}
	if len(merged.CategoriesOf(KindSpending)) != 3 {
		t.Fatalf("spending categories: %+v", merged.CategoriesOf(KindSpending))
	}
}

// "The application should be multi-lingual ... The choice should be saved.
// German should be the default language, if a choice has not been made and
// saved." - the backend stores an empty language until the user chooses,
// the frontend maps that to German (see frontend/src/i18n/index.ts).
func TestLanguageSetting(t *testing.T) {
	s, path := newTestService(t)
	if s.GetState().Settings.Language != "" {
		t.Fatalf("language before any choice = %q, want empty", s.GetState().Settings.Language)
	}
	raw, _ := os.ReadFile(path)
	if d, err := Decode(raw); err != nil || d.Settings.Language != "" {
		t.Fatalf("empty (unchosen) language must survive load: %v %+v", err, d.Settings)
	}
	st, err := s.SetLanguage("de")
	if err != nil {
		t.Fatal(err)
	}
	if st.Settings.Language != "de" {
		t.Fatalf("language = %q, want de", st.Settings.Language)
	}
	for _, bad := range []string{"", "german", "DE", "d1", "de_DE"} {
		if _, err := s.SetLanguage(bad); err == nil {
			t.Errorf("expected error for language %q", bad)
		}
	}
	if _, err := s.SetLanguage("pt-BR"); err != nil {
		t.Errorf("region codes should be accepted: %v", err)
	}
	s.SetLanguage("de")

	reloaded, err := NewService(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.GetState().Settings.Language != "de" {
		t.Fatal("language choice not persisted")
	}

	// Importing a file does not change the language of this installation.
	other, _ := newTestService(t)
	exportPath := filepath.Join(t.TempDir(), "x.json")
	other.ExportTo(exportPath)
	for _, mode := range []ImportMode{ImportMerge, ImportReplace} {
		res, err := reloaded.ImportData(exportPath, mode)
		if err != nil {
			t.Fatal(err)
		}
		if res.State.Settings.Language != "de" {
			t.Fatalf("import (%s) changed the language to %q", mode, res.State.Settings.Language)
		}
	}
}

// Errors carry a stable code and parameters so the UI can translate them.
func TestErrorsAreCoded(t *testing.T) {
	s, _ := newTestService(t)
	cat := mustCategory(t, s, KindSpending, "Housing")
	mustEntry(t, s, cat, "Rent", 1, PeriodMonthly)

	_, err := s.DeleteCategory(cat.ID)
	var coded *Error
	if !errors.As(err, &coded) {
		t.Fatalf("expected coded error, got %T: %v", err, err)
	}
	if coded.Code != ErrCategoryInUse || coded.Params["name"] != "Housing" || coded.Params["count"] != 1 {
		t.Fatalf("unexpected error: %+v", coded)
	}

	raw := MarshalError(err)
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("MarshalError produced invalid JSON: %v", err)
	}
	if decoded["code"] != ErrCategoryInUse {
		t.Fatalf("marshalled code = %v", decoded["code"])
	}
	if MarshalError(errors.New("plain")) != nil {
		t.Fatal("plain errors must fall back to the default handler")
	}
}

func TestLoadMissingFileGivesEmptyData(t *testing.T) {
	d, err := LoadFile(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if d.Version != CurrentVersion || len(d.Categories) != 0 || len(d.Entries) != 0 {
		t.Fatalf("unexpected data: %+v", d)
	}
}

// Version handling: unversioned files are migrated, newer files are rejected,
// broken references are rejected.
func TestDecodeVersions(t *testing.T) {
	unversioned := []byte(`{"categories":[{"id":"c1","name":"X","kind":"spending"}],"entries":[]}`)
	d, err := Decode(unversioned)
	if err != nil {
		t.Fatalf("migrating version 0: %v", err)
	}
	if d.Version != CurrentVersion || len(d.Categories) != 1 || d.Settings.Language != "" {
		t.Fatalf("migration result wrong: %+v", d)
	}

	v1 := []byte(`{"version":1,"categories":[{"id":"c1","name":"X","kind":"income"}],"entries":[{"id":"e1","categoryId":"c1","name":"Y","amountCents":100,"period":"yearly"}]}`)
	d, err = Decode(v1)
	if err != nil {
		t.Fatalf("migrating version 1: %v", err)
	}
	if d.Version != CurrentVersion || d.Settings.Language != "" || len(d.Entries) != 1 {
		t.Fatalf("v1 migration result wrong: %+v", d)
	}

	v2 := []byte(`{"version":2,"settings":{"language":"de"},"categories":[{"id":"c1","name":"X","kind":"spending"}],"entries":[{"id":"e1","categoryId":"c1","name":"Y","amountCents":1200,"period":"yearly"}]}`)
	d, err = Decode(v2)
	if err != nil {
		t.Fatalf("migrating version 2: %v", err)
	}
	e := d.Entries[0]
	if d.Version != 3 || d.Settings.Language != "de" || e.DueMonth != 0 || e.Paused || e.Notes != "" || d.Settings.SavingsGoalCents != 0 {
		t.Fatalf("v2 migration result wrong: %+v %+v", d.Settings, e)
	}

	newer := []byte(`{"version":999,"categories":[],"entries":[]}`)
	if _, err := Decode(newer); err == nil {
		t.Fatal("expected error for newer version")
	}

	dangling := []byte(`{"version":1,"categories":[],"entries":[{"id":"e1","categoryId":"missing","name":"X","amountCents":1,"period":"monthly"}]}`)
	if _, err := Decode(dangling); err == nil {
		t.Fatal("expected error for dangling category reference")
	}

	if _, err := Decode([]byte(`not json`)); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// "It should be possible to export and import the JSON file. When importing
// ask if the data should be added or if the current data should be replaced."
func TestExportImportReplaceAndMerge(t *testing.T) {
	src, _ := newTestService(t)
	housing := mustCategory(t, src, KindSpending, "Housing")
	mustEntry(t, src, housing, "Rent", 100000, PeriodMonthly)
	exportPath := filepath.Join(t.TempDir(), "export.json")
	if err := src.ExportTo(exportPath); err != nil {
		t.Fatal(err)
	}

	preview, err := src.PreviewImport(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Categories != 1 || preview.Entries != 1 || preview.Version != CurrentVersion {
		t.Fatalf("preview wrong: %+v", preview)
	}

	// Target has an equally named category plus another one.
	dst, _ := newTestService(t)
	dstHousing := mustCategory(t, dst, KindSpending, "housing")
	mustEntry(t, dst, dstHousing, "Power", 8000, PeriodMonthly)
	leisure := mustCategory(t, dst, KindSpending, "Leisure")
	mustEntry(t, dst, leisure, "Gym", 4000, PeriodMonthly)

	// Merge: "Rent" lands in the existing "housing" category (matched by
	// name, the ids differ), nothing is lost.
	res, err := dst.ImportData(exportPath, ImportMerge)
	if err != nil {
		t.Fatal(err)
	}
	st := res.State
	if len(st.Spending) != 2 {
		t.Fatalf("merge created duplicate categories: %+v", st.Spending)
	}
	if got := names(st.Spending[0].Entries); !equalStrings(got, []string{"Power", "Rent"}) {
		t.Fatalf("merged housing entries: %v", got)
	}
	if res.EntriesAdded != 1 || res.EntriesSkipped != 0 || res.CategoriesReused != 1 || res.CategoriesAdded != 0 {
		t.Fatalf("merge report wrong: %+v", res)
	}
	if err := dst.data.Validate(); err != nil {
		t.Fatalf("merged data invalid: %v", err)
	}
	// Importing the same file again skips the entry that already exists.
	res, err = dst.ImportData(exportPath, ImportMerge)
	if err != nil {
		t.Fatal(err)
	}
	if got := names(res.State.Spending[0].Entries); !equalStrings(got, []string{"Power", "Rent"}) {
		t.Fatalf("second merge duplicated entries: %v", got)
	}
	if res.EntriesAdded != 0 || res.EntriesSkipped != 1 {
		t.Fatalf("second merge report wrong: %+v", res)
	}

	// Replace: only the imported data remains.
	res, err = dst.ImportData(exportPath, ImportReplace)
	if err != nil {
		t.Fatal(err)
	}
	st = res.State
	if len(st.Spending) != 1 || !equalStrings(names(st.Spending[0].Entries), []string{"Rent"}) {
		t.Fatalf("replace result: %+v", st.Spending)
	}
	if res.EntriesAdded != 1 || res.CategoriesAdded != 1 || res.EntriesSkipped != 0 {
		t.Fatalf("replace report wrong: %+v", res)
	}

	// Invalid mode and invalid file are rejected without touching the data.
	if _, err := dst.ImportData(exportPath, ImportMode("append")); err == nil {
		t.Fatal("expected error for unknown mode")
	}
	bad := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(bad, []byte(`{"version":1,"entries":[{"id":"x"}]}`), 0o644)
	if _, err := dst.ImportData(bad, ImportReplace); err == nil {
		t.Fatal("expected error for invalid file")
	}
	if len(dst.GetState().Spending) != 1 {
		t.Fatal("failed import must not change the data")
	}
}

// Quarterly and half-yearly periods convert exactly like yearly ones.
func TestQuarterlyAndHalfYearly(t *testing.T) {
	q := Entry{AmountCents: 30000, Period: PeriodQuarterly}
	if q.MonthlyCents() != 10000 || q.YearlyCents() != 120000 {
		t.Fatalf("quarterly: %d / %d", q.MonthlyCents(), q.YearlyCents())
	}
	h := Entry{AmountCents: 60000, Period: PeriodHalfYearly}
	if h.MonthlyCents() != 10000 || h.YearlyCents() != 120000 {
		t.Fatalf("half-yearly: %d / %d", h.MonthlyCents(), h.YearlyCents())
	}
	odd := Entry{AmountCents: 10000, Period: PeriodQuarterly} // 33,333.. -> 33,33
	if odd.MonthlyCents() != 3333 {
		t.Fatalf("rounding: %d", odd.MonthlyCents())
	}
	for _, p := range []Period{PeriodMonthly, PeriodQuarterly, PeriodHalfYearly, PeriodYearly} {
		if !p.Valid() {
			t.Errorf("%s should be valid", p)
		}
	}
	// Non-monthly periods are saved up, only monthly ones go to the bank.
	s, _ := newTestService(t)
	cat := mustCategory(t, s, KindSpending, "Insurance")
	mustEntry(t, s, cat, "Car", 30000, PeriodQuarterly)
	mustEntry(t, s, cat, "Rent", 100000, PeriodMonthly)
	st := s.GetState().Stats
	if st.ToBankMonthlyCents != 100000 || st.ToSavingsMonthlyCents != 10000 || st.SpendingMonthlyCents != 110000 {
		t.Fatalf("stats: %+v", st)
	}
}

// Paused entries stay visible but count nowhere.
func TestPausedEntries(t *testing.T) {
	s, _ := newTestService(t)
	salary := mustCategory(t, s, KindIncome, "Salary")
	housing := mustCategory(t, s, KindSpending, "Housing")
	mustEntry(t, s, salary, "Job", 300000, PeriodMonthly)
	rent := mustEntry(t, s, housing, "Rent", 100000, PeriodMonthly)
	gym := mustEntry(t, s, housing, "Gym", 12000, PeriodYearly)

	if _, err := s.SetEntryPaused(rent.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetEntryPaused(gym.ID, true); err != nil {
		t.Fatal(err)
	}
	st := s.GetState()
	if len(st.Spending[0].Entries) != 2 {
		t.Fatal("paused entries must stay in the table")
	}
	if !st.Spending[0].Entries[0].Paused || st.Spending[0].Entries[0].MonthlyCents != 100000 {
		t.Fatalf("paused entry view wrong: %+v", st.Spending[0].Entries[0])
	}
	if st.Spending[0].MonthlyCents != 0 || st.Spending[0].YearlyCents != 0 {
		t.Fatalf("paused entries counted in subtotal: %+v", st.Spending[0])
	}
	if st.Stats.SpendingMonthlyCents != 0 || st.Stats.ToBankMonthlyCents != 0 || st.Stats.ToSavingsMonthlyCents != 0 || st.Stats.SaldoMonthlyCents != 300000 {
		t.Fatalf("paused entries counted in stats: %+v", st.Stats)
	}
	if st.Stats.PausedCount != 2 {
		t.Fatalf("paused count = %d", st.Stats.PausedCount)
	}
	st, _ = s.SetEntryPaused(rent.ID, false)
	if st.Stats.SpendingMonthlyCents != 100000 {
		t.Fatal("resumed entry not counted")
	}
	if _, err := s.SetEntryPaused("nope", true); err == nil {
		t.Fatal("expected error for unknown entry")
	}
}

// Timeline: what is due per month and how much the savings account holds.
func TestTimelineAndPeakBuffer(t *testing.T) {
	s, _ := newTestService(t)
	cat := mustCategory(t, s, KindSpending, "Insurance")
	// Yearly 1200 due in March: saved 100/month, paid in March.
	if _, err := s.AddEntry(EntryInput{CategoryID: cat.ID, Name: "Car", AmountCents: 120000, Period: PeriodYearly, DueMonth: 3}); err != nil {
		t.Fatal(err)
	}
	// Quarterly 300 due in January (and April, July, October).
	if _, err := s.AddEntry(EntryInput{CategoryID: cat.ID, Name: "Water", AmountCents: 30000, Period: PeriodQuarterly, DueMonth: 1}); err != nil {
		t.Fatal(err)
	}
	// Yearly without due month: counted in savings, but not in the timeline.
	mustEntry(t, s, cat, "Misc", 12000, PeriodYearly)
	// Paused yearly with due month: ignored completely.
	if _, err := s.AddEntry(EntryInput{CategoryID: cat.ID, Name: "Old", AmountCents: 99999, Period: PeriodYearly, DueMonth: 6, Paused: true}); err != nil {
		t.Fatal(err)
	}

	st := s.GetState().Stats
	tl := st.Timeline
	if tl[0].DueCents != 30000 || tl[2].DueCents != 120000 || tl[3].DueCents != 30000 || tl[5].DueCents != 0 {
		t.Fatalf("due wrong: %+v", tl)
	}
	// Car: balance at end of month t = 10000 * ((t-2) mod 12); Water: 10000 * (t mod 3).
	want := func(t int) int64 {
		car := int64(((t-2)%12+12)%12) * 10000
		water := int64(t%3) * 10000
		return car + water
	}
	for m := 0; m < 12; m++ {
		if tl[m].SavedCents != want(m) {
			t.Fatalf("month %d saved = %d, want %d", m+1, tl[m].SavedCents, want(m))
		}
	}
	// Peak: February = 110000 (car) + 10000 (water) = 120000.
	if st.PeakBufferCents != 120000 {
		t.Fatalf("peak buffer = %d", st.PeakBufferCents)
	}
	if st.UnscheduledCount != 1 {
		t.Fatalf("unscheduled = %d", st.UnscheduledCount)
	}
	if st.ToSavingsMonthlyCents != 10000+10000+1000 {
		t.Fatalf("to savings = %d", st.ToSavingsMonthlyCents)
	}
}

// Savings goal: remaining after goal and reachability.
func TestSavingsGoal(t *testing.T) {
	s, path := newTestService(t)
	salary := mustCategory(t, s, KindIncome, "Salary")
	housing := mustCategory(t, s, KindSpending, "Housing")
	mustEntry(t, s, salary, "Job", 300000, PeriodMonthly)
	mustEntry(t, s, housing, "Rent", 100000, PeriodMonthly)

	st, err := s.SetSavingsGoal(150000)
	if err != nil {
		t.Fatal(err)
	}
	if st.Stats.SavingsGoalCents != 150000 || st.Stats.RemainingAfterGoalCents != 50000 || !st.Stats.GoalReachable {
		t.Fatalf("goal stats: %+v", st.Stats)
	}
	st, _ = s.SetSavingsGoal(250000)
	if st.Stats.RemainingAfterGoalCents != -50000 || st.Stats.GoalReachable {
		t.Fatalf("unreachable goal stats: %+v", st.Stats)
	}
	if _, err := s.SetSavingsGoal(-1); err == nil {
		t.Fatal("expected error for negative goal")
	}
	reloaded, _ := LoadFile(path)
	if reloaded.Settings.SavingsGoalCents != 250000 {
		t.Fatal("goal not persisted")
	}
}

// Window geometry is stored and only written when it changed.
func TestWindowGeometry(t *testing.T) {
	s, path := newTestService(t)
	w := WindowGeometry{Width: 1500, Height: 950, X: 10, Y: 20}
	if err := s.SetWindow(w); err != nil {
		t.Fatal(err)
	}
	info1, _ := os.Stat(path)
	if err := s.SetWindow(w); err != nil {
		t.Fatal(err)
	}
	info2, _ := os.Stat(path)
	if info1.ModTime() != info2.ModTime() {
		t.Fatal("unchanged geometry must not rewrite the file")
	}
	reloaded, _ := LoadFile(path)
	if reloaded.Settings.Window != w {
		t.Fatalf("window not persisted: %+v", reloaded.Settings.Window)
	}
}

// Undo of deletions: entries and categories come back at their old position
// with their old id.
func TestRestoreEntryAndCategory(t *testing.T) {
	s, _ := newTestService(t)
	housing := mustCategory(t, s, KindSpending, "Housing")
	leisure := mustCategory(t, s, KindSpending, "Leisure")
	mustEntry(t, s, housing, "Rent", 100000, PeriodMonthly)
	power := mustEntry(t, s, housing, "Power", 8000, PeriodMonthly)
	mustEntry(t, s, housing, "Water", 3000, PeriodMonthly)

	if _, err := s.DeleteEntry(power.ID); err != nil {
		t.Fatal(err)
	}
	st, err := s.RestoreEntry(power, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := names(st.Spending[0].Entries); !equalStrings(got, []string{"Rent", "Power", "Water"}) {
		t.Fatalf("after restore: %v", got)
	}
	if st.Spending[0].Entries[1].ID != power.ID {
		t.Fatal("restored entry lost its id")
	}
	if _, err := s.RestoreEntry(power, 1); err == nil {
		t.Fatal("restoring an existing entry must fail")
	}

	// Category: delete the empty "Leisure" (index 1), restore at index 0.
	if _, err := s.DeleteCategory(leisure.ID); err != nil {
		t.Fatal(err)
	}
	st, err = s.RestoreCategory(leisure, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Spending) != 2 || st.Spending[0].ID != leisure.ID || st.Spending[0].Name != "Leisure" {
		t.Fatalf("category not restored at index 0: %+v", st.Spending)
	}
	if _, err := s.RestoreCategory(leisure, 0); err == nil {
		t.Fatal("restoring an existing category must fail")
	}
}

// Backups: automatic (rate limited), forced before import, list and restore.
func TestBackups(t *testing.T) {
	s, path := newTestService(t)
	clock := time.Date(2026, 1, 1, 12, 0, 0, 0, time.Local)
	s.now = func() time.Time { return clock }
	tick := func(d time.Duration) { clock = clock.Add(d) }

	cat := mustCategory(t, s, KindSpending, "Housing") // first edit -> backup of the empty file
	tick(time.Second)
	mustEntry(t, s, cat, "Rent", 100000, PeriodMonthly) // within the interval -> no backup
	list, err := s.ListBackups()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Categories != 0 {
		t.Fatalf("expected one backup of the empty file, got %+v", list)
	}
	tick(backupInterval)
	mustEntry(t, s, cat, "Power", 8000, PeriodMonthly) // interval passed -> backup with 1 category / 1 entry
	list, _ = s.ListBackups()
	if len(list) != 2 || list[0].Entries != 1 || list[0].Categories != 1 {
		t.Fatalf("expected a second backup with one entry, got %+v", list)
	}
	if !list[0].Time.After(list[1].Time) {
		t.Fatal("backups must be newest first")
	}

	// Import always creates a backup and reports it; restoring it undoes the import.
	other, _ := newTestService(t)
	oc := mustCategory(t, other, KindIncome, "Salary")
	mustEntry(t, other, oc, "Job", 300000, PeriodMonthly)
	exportPath := filepath.Join(t.TempDir(), "x.json")
	other.ExportTo(exportPath)
	tick(time.Second)
	res, err := s.ImportData(exportPath, ImportReplace)
	if err != nil {
		t.Fatal(err)
	}
	if res.BackupPath == "" || !isBackupPath(path, res.BackupPath) {
		t.Fatalf("import did not report its backup: %+v", res)
	}
	if len(res.State.Spending) != 0 {
		t.Fatal("replace import failed")
	}
	tick(time.Second)
	st, err := s.RestoreBackup(res.BackupPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Spending) != 1 || len(st.Spending[0].Entries) != 2 || len(st.Income) != 0 {
		t.Fatalf("restore did not undo the import: %+v", st)
	}
	// Only files inside the backup folder can be restored.
	if _, err := s.RestoreBackup(exportPath); err == nil {
		t.Fatal("restoring a file outside the backup folder must fail")
	}
	if _, err := s.RestoreBackup(filepath.Join(BackupDir(path), "missing.json")); err == nil {
		t.Fatal("restoring a missing backup must fail")
	}

	// Pruning keeps MaxBackups files.
	for i := 0; i < MaxBackups+5; i++ {
		tick(backupInterval)
		if _, err := s.SetCategoryCollapsed(cat.ID, i%2 == 0); err != nil {
			t.Fatal(err)
		}
	}
	list, _ = s.ListBackups()
	if len(list) != MaxBackups {
		t.Fatalf("expected %d backups after pruning, got %d", MaxBackups, len(list))
	}
}

// CSV export: header, separators and decimal marks per language.
func TestWriteCSV(t *testing.T) {
	d := NewData()
	d.Categories = []Category{
		{ID: "c1", Name: "Housing", Kind: KindSpending},
		{ID: "c2", Name: "Salary", Kind: KindIncome},
	}
	d.Entries = []Entry{
		{ID: "e1", CategoryID: "c1", Name: "Rent; big", AmountCents: 123456, Period: PeriodMonthly, Notes: "note"},
		{ID: "e2", CategoryID: "c1", Name: "Car", AmountCents: 60000, Period: PeriodYearly, DueMonth: 3, Paused: true},
		{ID: "e3", CategoryID: "c2", Name: "Job", AmountCents: 300000, Period: PeriodMonthly},
	}

	var de bytes.Buffer
	if err := WriteCSV(&de, d, "de"); err != nil {
		t.Fatal(err)
	}
	got := de.String()
	if !strings.HasPrefix(got, utf8BOM) {
		t.Fatal("missing UTF-8 BOM")
	}
	lines := strings.Split(strings.TrimSpace(strings.TrimPrefix(got, utf8BOM)), "\r\n")
	if len(lines) != 4 {
		t.Fatalf("expected header + 3 rows, got %d: %q", len(lines), got)
	}
	if lines[0] != "Art;Kategorie;Name;Zeitraum;Betrag;Pro Monat;Pro Jahr;Fälligkeitsmonat;Pausiert;Notizen" {
		t.Fatalf("header: %s", lines[0])
	}
	// Income first, then spending; German decimal comma; quoted field with separator.
	if lines[1] != "Einnahme;Salary;Job;monatlich;3000,00;3000,00;36000,00;;nein;" {
		t.Fatalf("row 1: %s", lines[1])
	}
	if lines[2] != `Ausgabe;Housing;"Rent; big";monatlich;1234,56;1234,56;14814,72;;nein;note` {
		t.Fatalf("row 2: %s", lines[2])
	}
	if lines[3] != "Ausgabe;Housing;Car;jährlich;600,00;50,00;600,00;3;ja;" {
		t.Fatalf("row 3: %s", lines[3])
	}

	var en bytes.Buffer
	if err := WriteCSV(&en, d, "en"); err != nil {
		t.Fatal(err)
	}
	enLines := strings.Split(strings.TrimSpace(strings.TrimPrefix(en.String(), utf8BOM)), "\r\n")
	if enLines[0] != "Kind,Category,Name,Period,Amount,Per month,Per year,Due month,Paused,Notes" {
		t.Fatalf("en header: %s", enLines[0])
	}
	if enLines[3] != "Spending,Housing,Car,yearly,600.00,50.00,600.00,3,yes," {
		t.Fatalf("en row 3: %s", enLines[3])
	}

	// Service level: writes the file in the current language.
	s, _ := newTestService(t)
	s.SetLanguage("de")
	c := mustCategory(t, s, KindSpending, "X")
	mustEntry(t, s, c, "Y", 100, PeriodMonthly)
	csvPath := filepath.Join(t.TempDir(), "out.csv")
	if err := s.ExportCSVTo(csvPath); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(csvPath)
	if !strings.Contains(string(raw), "Ausgabe;X;Y;monatlich;1,00") {
		t.Fatalf("csv file content: %q", raw)
	}
}

// Sample data only fills an empty planner and is valid in both languages.
func TestSampleData(t *testing.T) {
	for _, lang := range []string{"de", "en"} {
		d := SampleData(lang)
		if err := d.Validate(); err != nil {
			t.Fatalf("%s sample invalid: %v", lang, err)
		}
		if len(d.CategoriesOf(KindIncome)) == 0 || len(d.CategoriesOf(KindSpending)) < 3 || len(d.Entries) < 10 {
			t.Fatalf("%s sample too small: %d categories, %d entries", lang, len(d.Categories), len(d.Entries))
		}
	}
	if SampleData("de").Categories[0].Name == SampleData("en").Categories[0].Name {
		t.Fatal("sample data should be translated")
	}

	s, _ := newTestService(t)
	s.SetLanguage("de")
	res, err := s.LoadSampleData("de")
	if err != nil {
		t.Fatal(err)
	}
	if res.Entries != len(SampleData("de").Entries) || res.State.Settings.Language != "de" {
		t.Fatalf("sample result: %+v", res)
	}
	if st := res.State.Stats; st.PeakBufferCents <= 0 || st.ToBankMonthlyCents <= 0 {
		t.Fatalf("sample should produce a timeline and bank transfer: %+v", st)
	}
	if _, err := s.LoadSampleData("de"); err == nil {
		t.Fatal("sample data must not overwrite existing data")
	}
}

// The color scheme is an explicit choice saved in the settings; light is
// the default until one is made.
func TestThemeSetting(t *testing.T) {
	s, path := newTestService(t)
	if s.GetState().Settings.Theme != "" {
		t.Fatalf("theme before any choice = %q, want empty (light)", s.GetState().Settings.Theme)
	}
	for _, theme := range Themes {
		st, err := s.SetTheme(theme)
		if err != nil {
			t.Fatalf("SetTheme(%s): %v", theme, err)
		}
		if st.Settings.Theme != theme {
			t.Fatalf("theme = %q, want %q", st.Settings.Theme, theme)
		}
	}
	for _, bad := range []string{"", "blue", "Dark"} {
		if _, err := s.SetTheme(bad); err == nil {
			t.Errorf("expected error for theme %q", bad)
		}
	}
	reloaded, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Settings.Theme != "system" {
		t.Fatalf("theme not persisted: %q", reloaded.Settings.Theme)
	}
	if _, err := Decode([]byte(`{"version":3,"settings":{"theme":"purple"},"categories":[],"entries":[]}`)); err == nil {
		t.Fatal("unknown theme must be rejected when loading")
	}
}

// "Biggest levers": the five most expensive active spendings per year.
func TestTopSpendings(t *testing.T) {
	s, _ := newTestService(t)
	salary := mustCategory(t, s, KindIncome, "Salary")
	housing := mustCategory(t, s, KindSpending, "Housing")
	leisure := mustCategory(t, s, KindSpending, "Leisure")
	mustEntry(t, s, salary, "Job", 400000, PeriodMonthly)      // 4000/month
	mustEntry(t, s, housing, "Rent", 100000, PeriodMonthly)    // 12000/year
	mustEntry(t, s, housing, "Power", 8000, PeriodMonthly)     // 960/year
	mustEntry(t, s, leisure, "Holiday", 240000, PeriodYearly)  // 2400/year
	mustEntry(t, s, leisure, "Gym", 3000, PeriodMonthly)       // 360/year
	mustEntry(t, s, leisure, "Streaming", 1000, PeriodMonthly) // 120/year
	mustEntry(t, s, leisure, "Magazine", 6000, PeriodYearly)   // 60/year
	mustEntry(t, s, leisure, "Club", 6000, PeriodYearly)       // 60/year, same as Magazine
	paused := mustEntry(t, s, leisure, "Boat", 9999900, PeriodYearly)
	s.SetEntryPaused(paused.ID, true)

	st := s.GetState()
	top := st.Stats.TopSpendings
	if len(top) != TopEntryCount {
		t.Fatalf("expected %d levers, got %d", TopEntryCount, len(top))
	}
	got := []string{}
	for _, e := range top {
		got = append(got, e.Name)
	}
	// Ties (Magazine/Club at 60/year) are broken by name; Boat is paused.
	if !equalStrings(got, []string{"Rent", "Holiday", "Power", "Gym", "Streaming"}) {
		t.Fatalf("levers: %v", got)
	}
	if top[0].CategoryName != "Housing" || top[0].YearlyCents != 1200000 {
		t.Fatalf("first lever wrong: %+v", top[0])
	}
	// Shares: spending per month = 1000+80+200+30+10+5+5 = 1330; income 4000.
	if math.Abs(top[0].ShareOfSpending-100000.0/133000.0) > 1e-9 || math.Abs(top[0].ShareOfIncome-0.25) > 1e-9 {
		t.Fatalf("shares wrong: %+v", top[0])
	}
	// Category share of income: Housing 1080/4000.
	if math.Abs(st.Spending[0].ShareOfIncome-0.27) > 1e-9 {
		t.Fatalf("category share of income = %v", st.Spending[0].ShareOfIncome)
	}
	// Without income the share of income is 0, and an empty planner has no levers.
	empty, _ := newTestService(t)
	if len(empty.GetState().Stats.TopSpendings) != 0 {
		t.Fatal("empty planner must have no levers")
	}
}
