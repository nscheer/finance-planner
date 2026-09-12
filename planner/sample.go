package planner

import "strings"

// sampleEntry is one line of the example data set.
type sampleEntry struct {
	category    string
	name        string
	amountCents int64
	period      Period
	dueMonth    int
}

// SampleData returns a small, typical household plan in the given language.
// It is offered on an empty planner so the layout is understandable at once.
func SampleData(lang string) Data {
	de := strings.HasPrefix(lang, "de")
	pick := func(german, english string) string {
		if de {
			return german
		}
		return english
	}
	income := []sampleEntry{
		{pick("Gehalt", "Salary"), pick("Gehalt", "Salary"), 280000, PeriodMonthly, 0},
		{pick("Gehalt", "Salary"), pick("Weihnachtsgeld", "Christmas bonus"), 150000, PeriodYearly, 11},
		{pick("Sonstiges", "Other"), pick("Kindergeld", "Child benefit"), 25000, PeriodMonthly, 0},
	}
	spending := []sampleEntry{
		{pick("Wohnen", "Housing"), pick("Miete", "Rent"), 95000, PeriodMonthly, 0},
		{pick("Wohnen", "Housing"), pick("Strom", "Electricity"), 7500, PeriodMonthly, 0},
		{pick("Wohnen", "Housing"), pick("Internet", "Internet"), 3999, PeriodMonthly, 0},
		{pick("Versicherungen", "Insurance"), pick("Kfz-Versicherung", "Car insurance"), 62000, PeriodYearly, 1},
		{pick("Versicherungen", "Insurance"), pick("Haftpflicht", "Liability insurance"), 6500, PeriodYearly, 4},
		{pick("Versicherungen", "Insurance"), pick("Hausrat", "Household insurance"), 4800, PeriodHalfYearly, 3},
		{pick("Mobilität", "Mobility"), pick("Kfz-Steuer", "Car tax"), 18000, PeriodYearly, 7},
		{pick("Mobilität", "Mobility"), pick("Tanken", "Fuel"), 12000, PeriodMonthly, 0},
		{pick("Freizeit", "Leisure"), pick("Fitnessstudio", "Gym"), 2990, PeriodMonthly, 0},
		{pick("Freizeit", "Leisure"), pick("Streaming", "Streaming"), 1299, PeriodMonthly, 0},
		{pick("Freizeit", "Leisure"), pick("Urlaub", "Holiday"), 180000, PeriodYearly, 8},
		{pick("Rücklagen", "Reserves"), pick("Zahnzusatzversicherung", "Dental insurance"), 9000, PeriodQuarterly, 2},
	}
	d := NewData()
	add := func(kind Kind, items []sampleEntry) {
		ids := map[string]string{}
		for _, it := range items {
			id, ok := ids[it.category]
			if !ok {
				id = newID()
				ids[it.category] = id
				d.Categories = append(d.Categories, Category{ID: id, Name: it.category, Kind: kind})
			}
			d.Entries = append(d.Entries, Entry{
				ID: newID(), CategoryID: id, Name: it.name, AmountCents: it.amountCents,
				Period: it.period, DueMonth: it.dueMonth,
			})
		}
	}
	add(KindIncome, income)
	add(KindSpending, spending)
	return d
}
