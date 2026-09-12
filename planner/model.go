// Package planner contains the data model, calculations, persistence and the
// service that the Svelte frontend talks to.
//
// All monetary values are stored as integer cents to avoid floating point
// rounding problems. The frontend converts to/from a human readable € amount.
package planner

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// CurrentVersion is the version of the data.json structure written by this
// build. Bump it whenever the structure changes and add a migration step in
// migrate() (store.go).
//
// History:
//
//	1 - initial structure
//	2 - added "settings" (language)
const CurrentVersion = 2

// Kind distinguishes income from spending. Categories belong to exactly one
// kind, entries inherit the kind of their category.
type Kind string

const (
	KindIncome   Kind = "income"
	KindSpending Kind = "spending"
)

// Period is the "master" period of an entry: the period the user entered the
// amount for. The other period is derived (see calc.go).
type Period string

const (
	PeriodMonthly Period = "monthly"
	PeriodYearly  Period = "yearly"
)

// Category groups entries. The order of categories inside Data.Categories is
// the display order (per kind).
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind Kind   `json:"kind"`
	// Collapsed remembers whether the category is folded in the main view.
	Collapsed bool `json:"collapsed"`
}

// Entry is a single income or spending. The order of entries inside
// Data.Entries is the display order within a category.
type Entry struct {
	ID         string `json:"id"`
	CategoryID string `json:"categoryId"`
	Name       string `json:"name"`
	// AmountCents is the amount the user entered, in cents, for the Period.
	AmountCents int64  `json:"amountCents"`
	Period      Period `json:"period"`
}

// Settings holds user preferences that are stored together with the data.
type Settings struct {
	// Language is the UI language code chosen by the user, e.g. "en" or
	// "de". It stays empty until a choice was made; the frontend then uses
	// its default language (German).
	Language string `json:"language"`
}

// Data is the complete persisted state. It is serialised 1:1 to data.json.
type Data struct {
	Version    int        `json:"version"`
	Settings   Settings   `json:"settings"`
	Categories []Category `json:"categories"`
	Entries    []Entry    `json:"entries"`
}

// NewData returns an empty data set of the current version.
func NewData() Data {
	return Data{
		Version:    CurrentVersion,
		Categories: []Category{},
		Entries:    []Entry{},
	}
}

// ValidLanguage reports whether s looks like a language code ("en", "de",
// "pt-BR"). The list of actually supported languages lives in the frontend.
func ValidLanguage(s string) bool {
	if len(s) == 2 {
		return isLower(s)
	}
	if len(s) == 5 && s[2] == '-' {
		return isLower(s[:2]) && strings.ToUpper(s[3:]) == s[3:] && isLetters(s[3:])
	}
	return false
}

func isLower(s string) bool { return isLetters(s) && strings.ToLower(s) == s }

func isLetters(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return s != ""
}

// newID returns a random, URL-safe identifier.
func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand never fails on supported platforms
	}
	return hex.EncodeToString(b)
}

// Valid reports whether the kind is one of the known kinds.
func (k Kind) Valid() bool { return k == KindIncome || k == KindSpending }

// Valid reports whether the period is one of the known periods.
func (p Period) Valid() bool { return p == PeriodMonthly || p == PeriodYearly }

// Category returns the category with the given id, or nil.
func (d *Data) Category(id string) *Category {
	for i := range d.Categories {
		if d.Categories[i].ID == id {
			return &d.Categories[i]
		}
	}
	return nil
}

// Entry returns the entry with the given id, or nil.
func (d *Data) Entry(id string) *Entry {
	for i := range d.Entries {
		if d.Entries[i].ID == id {
			return &d.Entries[i]
		}
	}
	return nil
}

// CategoriesOf returns the categories of a kind in display order.
func (d *Data) CategoriesOf(kind Kind) []Category {
	var out []Category
	for _, c := range d.Categories {
		if c.Kind == kind {
			out = append(out, c)
		}
	}
	return out
}

// EntriesOf returns the entries of a category in display order.
func (d *Data) EntriesOf(categoryID string) []Entry {
	var out []Entry
	for _, e := range d.Entries {
		if e.CategoryID == categoryID {
			out = append(out, e)
		}
	}
	return out
}

// Validate checks referential integrity and enumerations. It is used after
// loading or importing a file so that corrupt data is rejected early.
func (d *Data) Validate() error {
	catIDs := map[string]bool{}
	for _, c := range d.Categories {
		if c.ID == "" {
			return fmt.Errorf("category %q has no id", c.Name)
		}
		if catIDs[c.ID] {
			return fmt.Errorf("duplicate category id %q", c.ID)
		}
		catIDs[c.ID] = true
		if strings.TrimSpace(c.Name) == "" {
			return fmt.Errorf("category %q has an empty name", c.ID)
		}
		if !c.Kind.Valid() {
			return fmt.Errorf("category %q has unknown kind %q", c.Name, c.Kind)
		}
	}
	if d.Settings.Language != "" && !ValidLanguage(d.Settings.Language) {
		return fmt.Errorf("invalid language %q in settings", d.Settings.Language)
	}
	entryIDs := map[string]bool{}
	for _, e := range d.Entries {
		if e.ID == "" {
			return fmt.Errorf("entry %q has no id", e.Name)
		}
		if entryIDs[e.ID] {
			return fmt.Errorf("duplicate entry id %q", e.ID)
		}
		entryIDs[e.ID] = true
		if strings.TrimSpace(e.Name) == "" {
			return fmt.Errorf("entry %q has an empty name", e.ID)
		}
		if !catIDs[e.CategoryID] {
			return fmt.Errorf("entry %q references unknown category %q", e.Name, e.CategoryID)
		}
		if !e.Period.Valid() {
			return fmt.Errorf("entry %q has unknown period %q", e.Name, e.Period)
		}
		if e.AmountCents < 0 {
			return fmt.Errorf("entry %q has a negative amount", e.Name)
		}
	}
	return nil
}
