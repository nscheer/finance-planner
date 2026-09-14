package planner

import (
	"math"
	"sort"
)

// MonthlyCents returns the monthly equivalent of an entry: the entered amount
// divided by the number of months per payment, rounded to cents.
func (e Entry) MonthlyCents() int64 {
	months := e.Period.Months()
	if months == 1 {
		return e.AmountCents
	}
	return int64(math.Round(float64(e.AmountCents) / float64(months)))
}

// YearlyCents returns the yearly equivalent of an entry (exact: the amount
// times the number of payments per year).
func (e Entry) YearlyCents() int64 {
	return e.AmountCents * int64(12/e.Period.Months())
}

// EntryView is an entry enriched with the derived values shown in the table.
type EntryView struct {
	Entry
	MonthlyCents int64 `json:"monthlyCents"`
	YearlyCents  int64 `json:"yearlyCents"`
}

// CategoryView is a category with its entries and subtotals. Paused entries
// are listed but do not count towards the subtotals.
type CategoryView struct {
	Category
	Entries      []EntryView `json:"entries"`
	MonthlyCents int64       `json:"monthlyCents"`
	YearlyCents  int64       `json:"yearlyCents"`
	// ShareOfIncome is the category's monthly total divided by the monthly
	// income (spending categories only, 0 without income).
	ShareOfIncome float64 `json:"shareOfIncome"`
}

// TopEntry is one of the "biggest levers": an active spending with one of
// the highest yearly costs.
type TopEntry struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	CategoryName    string  `json:"categoryName"`
	MonthlyCents    int64   `json:"monthlyCents"`
	YearlyCents     int64   `json:"yearlyCents"`
	ShareOfSpending float64 `json:"shareOfSpending"`
	ShareOfIncome   float64 `json:"shareOfIncome"`
}

// TopEntryCount is the length of Stats.TopSpendings.
const TopEntryCount = 5

// TimelineMonth describes one calendar month (index 0 = January) of the
// steady-state savings plan.
type TimelineMonth struct {
	// DueCents is the sum of non-monthly spendings that are paid in this month.
	DueCents int64 `json:"dueCents"`
	// SavedCents is the highest balance the savings account holds during the
	// month: after this month's contribution has arrived and before the
	// bills of the month are taken out. In a due month that is the full
	// amount saved for the bill.
	SavedCents int64 `json:"savedCents"`
}

// Stats is the content of the statistics box.
//
// The background: monthly spendings are paid from the bank account, so that
// amount has to be transferred to the bank account every month. Quarterly,
// half-yearly and yearly spendings are saved up on a savings account with
// 1/n of the amount every month, so that the money is available when the
// spending is due.
type Stats struct {
	IncomeMonthlyCents   int64 `json:"incomeMonthlyCents"`
	IncomeYearlyCents    int64 `json:"incomeYearlyCents"`
	SpendingMonthlyCents int64 `json:"spendingMonthlyCents"` // average cost per month
	SpendingYearlyCents  int64 `json:"spendingYearlyCents"`
	SaldoMonthlyCents    int64 `json:"saldoMonthlyCents"`
	SaldoYearlyCents     int64 `json:"saldoYearlyCents"`
	// ToBankMonthlyCents is the sum of all spendings entered per month.
	ToBankMonthlyCents int64 `json:"toBankMonthlyCents"`
	// ToSavingsMonthlyCents is the monthly share of all spendings that are
	// not paid monthly.
	ToSavingsMonthlyCents int64 `json:"toSavingsMonthlyCents"`

	// SavingsGoalCents is the amount the user wants to put aside per month.
	SavingsGoalCents int64 `json:"savingsGoalCents"`
	// RemainingAfterGoalCents is the monthly saldo minus the savings goal.
	RemainingAfterGoalCents int64 `json:"remainingAfterGoalCents"`
	GoalReachable           bool  `json:"goalReachable"`

	// Timeline shows, per calendar month, what is due and how much the
	// savings account holds. Only entries with a due month take part.
	Timeline [12]TimelineMonth `json:"timeline"`
	// PeakBufferCents is the highest balance of the year, i.e. the most the
	// savings account ever has to hold.
	PeakBufferCents int64 `json:"peakBufferCents"`
	// UnscheduledCount is the number of active non-monthly spendings without
	// a due month (not part of the timeline).
	UnscheduledCount int `json:"unscheduledCount"`
	// PausedCount is the number of paused entries (income and spending).
	PausedCount int `json:"pausedCount"`
	// TopSpendings lists the active spendings with the highest yearly cost.
	TopSpendings []TopEntry `json:"topSpendings"`
}

// State is everything the frontend needs to render the main view.
type State struct {
	Version int `json:"version"`
	// AppVersion is the version of the application (see version.go).
	AppVersion string         `json:"appVersion"`
	DataPath   string         `json:"dataPath"`
	Settings   Settings       `json:"settings"`
	Income     []CategoryView `json:"income"`
	Spending   []CategoryView `json:"spending"`
	Stats      Stats          `json:"stats"`
}

// BuildViews returns the categories of a kind with entries and subtotals.
func (d *Data) BuildViews(kind Kind) []CategoryView {
	views := []CategoryView{}
	for _, c := range d.CategoriesOf(kind) {
		v := CategoryView{Category: c, Entries: []EntryView{}}
		for _, e := range d.EntriesOf(c.ID) {
			ev := EntryView{Entry: e, MonthlyCents: e.MonthlyCents(), YearlyCents: e.YearlyCents()}
			if !e.Paused {
				v.MonthlyCents += ev.MonthlyCents
				v.YearlyCents += ev.YearlyCents
			}
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
		if e.Paused {
			s.PausedCount++
			continue
		}
		switch kinds[e.CategoryID] {
		case KindIncome:
			s.IncomeMonthlyCents += e.MonthlyCents()
			s.IncomeYearlyCents += e.YearlyCents()
		case KindSpending:
			s.SpendingMonthlyCents += e.MonthlyCents()
			s.SpendingYearlyCents += e.YearlyCents()
			if e.Period == PeriodMonthly {
				s.ToBankMonthlyCents += e.MonthlyCents()
				continue
			}
			s.ToSavingsMonthlyCents += e.MonthlyCents()
			if e.DueMonth == 0 {
				s.UnscheduledCount++
				continue
			}
			addToTimeline(&s.Timeline, e)
		}
	}
	s.SaldoMonthlyCents = s.IncomeMonthlyCents - s.SpendingMonthlyCents
	s.SaldoYearlyCents = s.IncomeYearlyCents - s.SpendingYearlyCents
	s.TopSpendings = d.topSpendings(kinds, s.SpendingMonthlyCents, s.IncomeMonthlyCents)

	s.SavingsGoalCents = d.Settings.SavingsGoalCents
	s.RemainingAfterGoalCents = s.SaldoMonthlyCents - s.SavingsGoalCents
	s.GoalReachable = s.RemainingAfterGoalCents >= 0

	for _, m := range s.Timeline {
		if m.SavedCents > s.PeakBufferCents {
			s.PeakBufferCents = m.SavedCents
		}
	}
	return s
}

// topSpendings returns the TopEntryCount active spendings with the highest
// yearly cost (ties broken by name), with their shares of all spending and
// of the income.
func (d *Data) topSpendings(kinds map[string]Kind, spendingMonthly, incomeMonthly int64) []TopEntry {
	names := map[string]string{}
	for _, c := range d.Categories {
		names[c.ID] = c.Name
	}
	top := []TopEntry{}
	for _, e := range d.Entries {
		if e.Paused || kinds[e.CategoryID] != KindSpending {
			continue
		}
		t := TopEntry{ID: e.ID, Name: e.Name, CategoryName: names[e.CategoryID], MonthlyCents: e.MonthlyCents(), YearlyCents: e.YearlyCents()}
		if spendingMonthly > 0 {
			t.ShareOfSpending = float64(t.MonthlyCents) / float64(spendingMonthly)
		}
		if incomeMonthly > 0 {
			t.ShareOfIncome = float64(t.MonthlyCents) / float64(incomeMonthly)
		}
		top = append(top, t)
	}
	sort.SliceStable(top, func(i, j int) bool {
		if top[i].YearlyCents != top[j].YearlyCents {
			return top[i].YearlyCents > top[j].YearlyCents
		}
		return top[i].Name < top[j].Name
	})
	if len(top) > TopEntryCount {
		top = top[:TopEntryCount]
	}
	return top
}

// addToTimeline adds a scheduled non-monthly spending to the timeline.
//
// Every month 1/n of the amount is put aside (n = months per payment) and in
// a due month the bill is taken out again. What the month shows is the high
// point of that month: the balance once this month's instalment has arrived
// and before the bill is paid. Counting from the month after the last
// payment, that is monthly × (k + 1) with k = (t − 1 − due) mod n, so a due
// month shows monthly × n, the full amount saved for the bill, which is
// where the line meets the bar. (The month-end balance would be the low
// point instead and would never show the money that is actually there.)
func addToTimeline(tl *[12]TimelineMonth, e Entry) {
	n := e.Period.Months()
	monthly := e.MonthlyCents()
	due := e.DueMonth - 1 // 0-based
	for t := 0; t < 12; t++ {
		k := ((t-1-due)%n + n) % n // months since the instalment run started
		tl[t].SavedCents += monthly * int64(k+1)
		if ((t-due)%n+n)%n == 0 {
			tl[t].DueCents += e.AmountCents
		}
	}
}

// BuildState assembles the full frontend state.
func (d *Data) BuildState(dataPath string) State {
	stats := d.ComputeStats()
	spending := d.BuildViews(KindSpending)
	if stats.IncomeMonthlyCents > 0 {
		for i := range spending {
			spending[i].ShareOfIncome = float64(spending[i].MonthlyCents) / float64(stats.IncomeMonthlyCents)
		}
	}
	return State{
		Version:    d.Version,
		AppVersion: AppVersion,
		DataPath:   dataPath,
		Settings:   d.Settings,
		Income:     d.BuildViews(KindIncome),
		Spending:   spending,
		Stats:      stats,
	}
}
