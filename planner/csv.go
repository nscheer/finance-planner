package planner

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// utf8BOM marks the file as UTF-8 for spreadsheet programs.
const utf8BOM = "\xEF\xBB\xBF"

// WriteCSV writes all entries as a flat table for spreadsheets. The language
// decides the separator and decimal mark: German uses ";" and a decimal
// comma (what Excel expects there), everything else "," and a decimal point.
// A UTF-8 byte order mark is written so that Excel detects the encoding.
func WriteCSV(w io.Writer, d Data, lang string) error {
	german := strings.HasPrefix(lang, "de")
	if _, err := io.WriteString(w, utf8BOM); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if german {
		cw.Comma = ';'
	}
	cw.UseCRLF = true

	header := []string{"Kind", "Category", "Name", "Period", "Amount", "Per month", "Per year", "Due month", "Paused", "Notes"}
	if german {
		header = []string{"Art", "Kategorie", "Name", "Zeitraum", "Betrag", "Pro Monat", "Pro Jahr", "Fälligkeitsmonat", "Pausiert", "Notizen"}
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	money := func(cents int64) string {
		s := strconv.FormatInt(cents, 10)
		neg := strings.HasPrefix(s, "-")
		s = strings.TrimPrefix(s, "-")
		for len(s) < 3 {
			s = "0" + s
		}
		mark := "."
		if german {
			mark = ","
		}
		out := s[:len(s)-2] + mark + s[len(s)-2:]
		if neg {
			out = "-" + out
		}
		return out
	}
	yesNo := func(b bool) string {
		if german {
			if b {
				return "ja"
			}
			return "nein"
		}
		if b {
			return "yes"
		}
		return "no"
	}
	for _, kind := range []Kind{KindIncome, KindSpending} {
		for _, c := range d.CategoriesOf(kind) {
			for _, e := range d.EntriesOf(c.ID) {
				due := ""
				if e.DueMonth > 0 {
					due = strconv.Itoa(e.DueMonth)
				}
				row := []string{
					kindLabel(kind, german), c.Name, e.Name, periodLabel(e.Period, german),
					money(e.AmountCents), money(e.MonthlyCents()), money(e.YearlyCents()),
					due, yesNo(e.Paused), e.Notes,
				}
				if err := cw.Write(row); err != nil {
					return err
				}
			}
		}
	}
	cw.Flush()
	return cw.Error()
}

func kindLabel(k Kind, german bool) string {
	switch {
	case k == KindIncome && german:
		return "Einnahme"
	case k == KindIncome:
		return "Income"
	case german:
		return "Ausgabe"
	default:
		return "Spending"
	}
}

func periodLabel(p Period, german bool) string {
	labels := map[Period][2]string{
		PeriodMonthly:    {"monthly", "monatlich"},
		PeriodQuarterly:  {"quarterly", "vierteljährlich"},
		PeriodHalfYearly: {"half-yearly", "halbjährlich"},
		PeriodYearly:     {"yearly", "jährlich"},
	}
	l, ok := labels[p]
	if !ok {
		return fmt.Sprint(p)
	}
	if german {
		return l[1]
	}
	return l[0]
}
