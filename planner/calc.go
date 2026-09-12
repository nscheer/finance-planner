package planner

import "math"

// MonthlyCents returns the monthly equivalent of an entry: the entered amount
// for monthly entries, 1/12 of the entered amount (rounded to cents) for
// yearly entries.
func (e Entry) MonthlyCents() int64 {
	if e.Period == PeriodYearly {
		return int64(math.Round(float64(e.AmountCents) / 12))
	}
	return e.AmountCents
}

// YearlyCents returns the yearly equivalent of an entry: 12 times the entered
// amount for monthly entries, the entered amount for yearly entries.
func (e Entry) YearlyCents() int64 {
	if e.Period == PeriodYearly {
		return e.AmountCents
	}
	return e.AmountCents * 12
}

// EntryView is an entry enriched with the derived values shown in the table.
type EntryView struct {
	Entry
	MonthlyCents int64 `json:"monthlyCents"`
	YearlyCents  int64 `json:"yearlyCents"`
}

// CategoryView is a category with its entries and subtotals.
type CategoryView struct {
	Category
	Entries      []EntryView `json:"entries"`
	MonthlyCents int64       `json:"monthlyCents"`
	YearlyCents  int64       `json:"yearlyCents"`
}

// Stats is the content of the statistics box.
//
// The background: monthly spendings are paid from the bank account, so that
// amount has to be transferred to the bank account every month. Yearly
// spendings are saved up on a savings account with 1/12 of the amount every
// month, so that the money is available when the spending is due.
type Stats struct {
	IncomeMonthlyCents   int64 `json:"incomeMonthlyCents"`
	IncomeYearlyCents    int64 `json:"incomeYearlyCents"`
	SpendingMonthlyCents int64 `json:"spendingMonthlyCents"` // average cost per month
	SpendingYearlyCents  int64 `json:"spendingYearlyCents"`
	SaldoMonthlyCents    int64 `json:"saldoMonthlyCents"`
	SaldoYearlyCents     int64 `json:"saldoYearlyCents"`
	// ToBankMonthlyCents is the sum of all spendings entered per month.
	ToBankMonthlyCents int64 `json:"toBankMonthlyCents"`
	// ToSavingsMonthlyCents is the sum of 1/12 of all spendings entered per year.
	ToSavingsMonthlyCents int64 `json:"toSavingsMonthlyCents"`
}

// State is everything the frontend needs to render the main view.
type State struct {
	Version  int            `json:"version"`
	DataPath string         `json:"dataPath"`
	Income   []CategoryView `json:"income"`
	Spending []CategoryView `json:"spending"`
	Stats    Stats          `json:"stats"`
}

// BuildViews returns the categories of a kind with entries and subtotals.
func (d *Data) BuildViews(kind Kind) []CategoryView {
	views := []CategoryView{}
	for _, c := range d.CategoriesOf(kind) {
		v := CategoryView{Category: c, Entries: []EntryView{}}
		for _, e := range d.EntriesOf(c.ID) {
			ev := EntryView{Entry: e, MonthlyCents: e.MonthlyCents(), YearlyCents: e.YearlyCents()}
			v.MonthlyCents += ev.MonthlyCents
			v.YearlyCents += ev.YearlyCents
			v.Entries = append(v.Entries, ev)
		}
		views = append(views, v)
	}
	return views
}

// ComputeStats calculates the statistics box values.
func (d *Data) ComputeStats() Stats {
	var s Stats
	kinds := map[string]Kind{}
	for _, c := range d.Categories {
		kinds[c.ID] = c.Kind
	}
	for _, e := range d.Entries {
		switch kinds[e.CategoryID] {
		case KindIncome:
			s.IncomeMonthlyCents += e.MonthlyCents()
			s.IncomeYearlyCents += e.YearlyCents()
		case KindSpending:
			s.SpendingMonthlyCents += e.MonthlyCents()
			s.SpendingYearlyCents += e.YearlyCents()
			if e.Period == PeriodYearly {
				s.ToSavingsMonthlyCents += e.MonthlyCents()
			} else {
				s.ToBankMonthlyCents += e.MonthlyCents()
			}
		}
	}
	s.SaldoMonthlyCents = s.IncomeMonthlyCents - s.SpendingMonthlyCents
	s.SaldoYearlyCents = s.IncomeYearlyCents - s.SpendingYearlyCents
	return s
}

// BuildState assembles the full frontend state.
func (d *Data) BuildState(dataPath string) State {
	return State{
		Version:  d.Version,
		DataPath: dataPath,
		Income:   d.BuildViews(KindIncome),
		Spending: d.BuildViews(KindSpending),
		Stats:    d.ComputeStats(),
	}
}
