package planner

// Test cases against project.md. Each test names the requirement it covers.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
	st, err := s.AddEntry(cat.ID, name, cents, period)
	if err != nil {
		t.Fatalf("AddEntry(%s): %v", name, err)
	}
	_ = st
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
	if _, err := s.AddEntry("does-not-exist", "Rent", 100000, PeriodMonthly); err == nil {
		t.Fatal("expected error when adding an entry without a category")
	}
	cat := mustCategory(t, s, KindSpending, "Housing")
	if _, err := s.AddEntry(cat.ID, "Rent", 100000, PeriodMonthly); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// "Each spending has a name, an amount in € and a category."
func TestEntryValidation(t *testing.T) {
	s, _ := newTestService(t)
	cat := mustCategory(t, s, KindSpending, "Housing")
	cases := []struct {
		name   string
		cents  int64
		period Period
	}{
		{"", 100, PeriodMonthly},
		{"   ", 100, PeriodMonthly},
		{"Rent", 0, PeriodMonthly},
		{"Rent", -5, PeriodMonthly},
		{"Rent", 100, Period("weekly")},
	}
	for _, c := range cases {
		if _, err := s.AddEntry(cat.ID, c.name, c.cents, c.period); err == nil {
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
	want := Stats{
		IncomeMonthlyCents:    310000,
		IncomeYearlyCents:     3720000,
		SpendingMonthlyCents:  125000,
		SpendingYearlyCents:   1500000,
		SaldoMonthlyCents:     185000,
		SaldoYearlyCents:      2220000,
		ToBankMonthlyCents:    120000,
		ToSavingsMonthlyCents: 5000,
	}
	if st != want {
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

	st, err := s.UpdateEntry(rent.ID, leisure.ID, "Rent (new)", 1200000, PeriodYearly)
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
	if _, err := s.UpdateEntry(rent.ID, leisure.ID, "", 1, PeriodMonthly); err == nil {
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
	if d.Version != CurrentVersion || len(d.Categories) != 1 {
		t.Fatalf("migration result wrong: %+v", d)
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

	// Merge: "Rent" lands in the existing "housing" category, nothing is lost.
	st, err := dst.ImportData(exportPath, ImportMerge)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Spending) != 2 {
		t.Fatalf("merge created duplicate categories: %+v", st.Spending)
	}
	if got := names(st.Spending[0].Entries); !equalStrings(got, []string{"Power", "Rent"}) {
		t.Fatalf("merged housing entries: %v", got)
	}
	if err := dst.data.Validate(); err != nil {
		t.Fatalf("merged data invalid: %v", err)
	}
	// Merged entries get new ids, so importing the same file twice works.
	if _, err := dst.ImportData(exportPath, ImportMerge); err != nil {
		t.Fatal(err)
	}
	if got := names(dst.GetState().Spending[0].Entries); !equalStrings(got, []string{"Power", "Rent", "Rent"}) {
		t.Fatalf("second merge: %v", got)
	}

	// Replace: only the imported data remains.
	st, err = dst.ImportData(exportPath, ImportReplace)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Spending) != 1 || !equalStrings(names(st.Spending[0].Entries), []string{"Rent"}) {
		t.Fatalf("replace result: %+v", st.Spending)
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
