package planner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ImportMode tells ImportData what to do with the current data.
type ImportMode string

const (
	// ImportReplace throws the current data away and uses the imported data.
	ImportReplace ImportMode = "replace"
	// ImportMerge adds the imported categories and entries to the current
	// data. Categories with the same name and kind are reused.
	ImportMerge ImportMode = "merge"
)

// ImportPreview describes a file selected for import, so that the UI can ask
// the user whether to merge or replace before anything is changed.
type ImportPreview struct {
	// Path is empty if the user cancelled the file dialog.
	Path       string `json:"path"`
	Version    int    `json:"version"`
	Categories int    `json:"categories"`
	Entries    int    `json:"entries"`
}

// ImportResult is returned by ImportData: the new state plus a report of
// what the import did, shown to the user in a notification.
type ImportResult struct {
	State            State `json:"state"`
	CategoriesAdded  int   `json:"categoriesAdded"`
	CategoriesReused int   `json:"categoriesReused"`
	EntriesAdded     int   `json:"entriesAdded"`
	EntriesSkipped   int   `json:"entriesSkipped"`
}

// Service is the Wails service used by the frontend. Every mutating method
// persists the data immediately and returns the new State, so the frontend
// simply replaces its state with the result.
type Service struct {
	mu   sync.Mutex
	path string
	data Data
}

// NewService creates a service that persists to the given file. The file is
// loaded (or created empty) on construction.
func NewService(path string) (*Service, error) {
	d, err := LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", path, err)
	}
	s := &Service{path: path, data: d}
	// Create the file right away so the user can see where the data lives.
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := s.save(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// ServiceName is shown in the Wails logs.
func (s *Service) ServiceName() string { return "PlannerService" }

// GetState returns the current state for rendering.
func (s *Service) GetState() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state()
}

// state must be called with the lock held.
func (s *Service) state() State { return s.data.BuildState(s.path) }

// save persists the data. Must be called with the lock held.
func (s *Service) save() error {
	if err := SaveFile(s.path, s.data); err != nil {
		return fmt.Errorf("saving %s: %w", s.path, err)
	}
	return nil
}

// mutate runs fn under the lock, saves on success and returns the new state.
func (s *Service) mutate(fn func() error) (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fn(); err != nil {
		return s.state(), err
	}
	if err := s.save(); err != nil {
		return s.state(), err
	}
	return s.state(), nil
}

// ---- categories -----------------------------------------------------------

// AddCategory appends a new category of the given kind.
func (s *Service) AddCategory(kind Kind, name string) (State, error) {
	return s.mutate(func() error {
		name = strings.TrimSpace(name)
		if !kind.Valid() {
			return newError(ErrKindUnknown, "kind", kind)
		}
		if name == "" {
			return newError(ErrCategoryNameEmpty)
		}
		if s.findCategoryByName(kind, name) != nil {
			return newError(ErrCategoryExists, "kind", kind, "name", name)
		}
		s.data.Categories = append(s.data.Categories, Category{ID: newID(), Name: name, Kind: kind})
		return nil
	})
}

// RenameCategory changes the name of a category.
func (s *Service) RenameCategory(id, name string) (State, error) {
	return s.mutate(func() error {
		name = strings.TrimSpace(name)
		c := s.data.Category(id)
		if c == nil {
			return newError(ErrCategoryNotFound)
		}
		if name == "" {
			return newError(ErrCategoryNameEmpty)
		}
		if other := s.findCategoryByName(c.Kind, name); other != nil && other.ID != id {
			return newError(ErrCategoryExists, "kind", c.Kind, "name", name)
		}
		c.Name = name
		return nil
	})
}

// DeleteCategory removes a category. It fails if any entry still uses it.
func (s *Service) DeleteCategory(id string) (State, error) {
	return s.mutate(func() error {
		c := s.data.Category(id)
		if c == nil {
			return newError(ErrCategoryNotFound)
		}
		if n := len(s.data.EntriesOf(id)); n > 0 {
			return newError(ErrCategoryInUse, "name", c.Name, "count", n)
		}
		s.data.Categories = removeCategory(s.data.Categories, id)
		return nil
	})
}

// SetCategoryCollapsed folds or unfolds one category.
func (s *Service) SetCategoryCollapsed(id string, collapsed bool) (State, error) {
	return s.mutate(func() error {
		c := s.data.Category(id)
		if c == nil {
			return newError(ErrCategoryNotFound)
		}
		c.Collapsed = collapsed
		return nil
	})
}

// SetAllCollapsed folds or unfolds all categories of one kind ("expand all"
// / "collapse all" at the top of the income and the spending block).
func (s *Service) SetAllCollapsed(kind Kind, collapsed bool) (State, error) {
	return s.mutate(func() error {
		if !kind.Valid() {
			return newError(ErrKindUnknown, "kind", kind)
		}
		for i := range s.data.Categories {
			if s.data.Categories[i].Kind == kind {
				s.data.Categories[i].Collapsed = collapsed
			}
		}
		return nil
	})
}

// MoveCategory moves a category to position toIndex among the categories of
// its own kind (0 = first).
func (s *Service) MoveCategory(id string, toIndex int) (State, error) {
	return s.mutate(func() error {
		c := s.data.Category(id)
		if c == nil {
			return newError(ErrCategoryNotFound)
		}
		moved := *c
		rest := removeCategory(s.data.Categories, id)
		// Find the global slice position of the toIndex-th category of the
		// same kind; append after the last one if toIndex is past the end.
		insertAt := len(rest)
		seen := 0
		for i, other := range rest {
			if other.Kind != moved.Kind {
				continue
			}
			if seen == toIndex {
				insertAt = i
				break
			}
			seen++
			insertAt = i + 1
		}
		s.data.Categories = insertCategory(rest, insertAt, moved)
		return nil
	})
}

// ---- entries --------------------------------------------------------------

// AddEntry appends a new entry to a category.
func (s *Service) AddEntry(categoryID, name string, amountCents int64, period Period) (State, error) {
	return s.mutate(func() error {
		if err := s.validateEntry(categoryID, &name, amountCents, period); err != nil {
			return err
		}
		s.data.Entries = append(s.data.Entries, Entry{
			ID: newID(), CategoryID: categoryID, Name: name, AmountCents: amountCents, Period: period,
		})
		return nil
	})
}

// UpdateEntry changes all editable fields of an entry. When the category
// changes, the entry is appended at the end of the new category.
func (s *Service) UpdateEntry(id, categoryID, name string, amountCents int64, period Period) (State, error) {
	return s.mutate(func() error {
		e := s.data.Entry(id)
		if e == nil {
			return newError(ErrEntryNotFound)
		}
		if err := s.validateEntry(categoryID, &name, amountCents, period); err != nil {
			return err
		}
		e.Name, e.AmountCents, e.Period = name, amountCents, period
		if e.CategoryID != categoryID {
			moved := *e
			moved.CategoryID = categoryID
			s.data.Entries = append(removeEntry(s.data.Entries, id), moved)
		}
		return nil
	})
}

// DeleteEntry removes an entry.
func (s *Service) DeleteEntry(id string) (State, error) {
	return s.mutate(func() error {
		if s.data.Entry(id) == nil {
			return newError(ErrEntryNotFound)
		}
		s.data.Entries = removeEntry(s.data.Entries, id)
		return nil
	})
}

// MoveEntry moves an entry to position toIndex (0 = first) inside the target
// category, which may be the entry's current category (reorder) or another
// category of the same kind (re-categorise).
func (s *Service) MoveEntry(id, targetCategoryID string, toIndex int) (State, error) {
	return s.mutate(func() error {
		e := s.data.Entry(id)
		if e == nil {
			return newError(ErrEntryNotFound)
		}
		from := s.data.Category(e.CategoryID)
		to := s.data.Category(targetCategoryID)
		if to == nil {
			return newError(ErrCategoryNotFound)
		}
		if from != nil && from.Kind != to.Kind {
			return newError(ErrEntryKindMismatch, "from", from.Kind, "to", to.Kind)
		}
		moved := *e
		moved.CategoryID = targetCategoryID
		rest := removeEntry(s.data.Entries, id)
		insertAt := len(rest)
		seen := 0
		for i, other := range rest {
			if other.CategoryID != targetCategoryID {
				continue
			}
			if seen == toIndex {
				insertAt = i
				break
			}
			seen++
			insertAt = i + 1
		}
		s.data.Entries = insertEntry(rest, insertAt, moved)
		return nil
	})
}

// ---- settings -------------------------------------------------------------

// SetLanguage stores the UI language chosen in the language dropdown.
func (s *Service) SetLanguage(language string) (State, error) {
	return s.mutate(func() error {
		if !ValidLanguage(language) {
			return newError(ErrLanguageInvalid, "language", language)
		}
		s.data.Settings.Language = language
		return nil
	})
}

// ---- import / export ------------------------------------------------------

// ExportTo writes the current data to the given path.
func (s *Service) ExportTo(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return SaveFile(path, s.data)
}

// PreviewImport reads a file and reports what it contains without changing
// anything.
func (s *Service) PreviewImport(path string) (ImportPreview, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ImportPreview{}, err
	}
	d, err := Decode(raw)
	if err != nil {
		return ImportPreview{}, newError(ErrImportInvalidFile, "file", filepath.Base(path), "detail", err.Error())
	}
	return ImportPreview{Path: path, Version: d.Version, Categories: len(d.Categories), Entries: len(d.Entries)}, nil
}

// ImportData imports a file, either replacing the current data or merging
// it into the current data (see Merge for the rules).
func (s *Service) ImportData(path string, mode ImportMode) (ImportResult, error) {
	var result ImportResult
	st, err := s.mutate(func() error {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		imported, err := Decode(raw)
		if err != nil {
			return newError(ErrImportInvalidFile, "file", filepath.Base(path), "detail", err.Error())
		}
		switch mode {
		case ImportReplace:
			// The language is a preference of this installation, not of the file.
			imported.Settings = s.data.Settings
			s.data = imported
			result = ImportResult{CategoriesAdded: len(imported.Categories), EntriesAdded: len(imported.Entries)}
		case ImportMerge:
			s.data, result = Merge(s.data, imported)
		default:
			return newError(ErrImportModeUnknown, "mode", mode)
		}
		return nil
	})
	result.State = st
	return result, err
}

// Merge adds the categories and entries of extra to base.
//
// Categories are matched by id first (same id and kind), then by kind and
// name (case-insensitive); unmatched categories are added and keep their id
// when it is free, so that later imports of the same file match by id.
// Entries whose id already exists are skipped, so importing overlapping
// files never creates duplicates; all other entries are added.
func Merge(base, extra Data) (Data, ImportResult) {
	var report ImportResult
	out := Data{Version: CurrentVersion, Settings: base.Settings}
	out.Categories = append(out.Categories, base.Categories...)
	out.Entries = append(out.Entries, base.Entries...)

	catMap := map[string]string{} // extra category id -> merged category id
	for _, c := range extra.Categories {
		found := ""
		if existing := out.Category(c.ID); existing != nil && existing.Kind == c.Kind {
			found = existing.ID
		} else {
			for _, existing := range out.Categories {
				if existing.Kind == c.Kind && strings.EqualFold(existing.Name, c.Name) {
					found = existing.ID
					break
				}
			}
		}
		if found != "" {
			report.CategoriesReused++
		} else {
			found = c.ID
			if found == "" || out.Category(found) != nil {
				found = newID()
			}
			out.Categories = append(out.Categories, Category{ID: found, Name: c.Name, Kind: c.Kind, Collapsed: c.Collapsed})
			report.CategoriesAdded++
		}
		catMap[c.ID] = found
	}
	for _, e := range extra.Entries {
		target, ok := catMap[e.CategoryID]
		if !ok {
			continue // Decode() guarantees this can't happen, but stay safe
		}
		if out.Entry(e.ID) != nil {
			report.EntriesSkipped++
			continue
		}
		if e.ID == "" {
			e.ID = newID()
		}
		e.CategoryID = target
		out.Entries = append(out.Entries, e)
		report.EntriesAdded++
	}
	return out, report
}

// ---- dialogs (need a running Wails application) ---------------------------

// ExportData asks the user for a target file and writes the data to it.
// It returns the chosen path, or "" if the user cancelled.
func (s *Service) ExportData() (string, error) {
	path, err := application.Get().Dialog.SaveFile().
		SetMessage("Export planner data").
		SetFilename("finance-planner-export.json").
		AddFilter("JSON files", "*.json").
		PromptForSingleSelection()
	if err != nil || path == "" {
		return "", err
	}
	if filepath.Ext(path) == "" {
		path += ".json"
	}
	return path, s.ExportTo(path)
}

// ChooseImportFile asks the user for a file and returns a preview of its
// content. The preview's Path is "" if the user cancelled.
func (s *Service) ChooseImportFile() (ImportPreview, error) {
	path, err := application.Get().Dialog.OpenFile().
		SetTitle("Import planner data").
		AddFilter("JSON files", "*.json").
		PromptForSingleSelection()
	if err != nil || path == "" {
		return ImportPreview{}, err
	}
	return s.PreviewImport(path)
}

// ---- helpers --------------------------------------------------------------

func (s *Service) findCategoryByName(kind Kind, name string) *Category {
	for i := range s.data.Categories {
		c := &s.data.Categories[i]
		if c.Kind == kind && strings.EqualFold(c.Name, name) {
			return c
		}
	}
	return nil
}

func (s *Service) validateEntry(categoryID string, name *string, amountCents int64, period Period) error {
	*name = strings.TrimSpace(*name)
	if s.data.Category(categoryID) == nil {
		return newError(ErrEntryCategoryRequired)
	}
	if *name == "" {
		return newError(ErrEntryNameEmpty)
	}
	if amountCents <= 0 {
		return newError(ErrEntryAmountPositive)
	}
	if !period.Valid() {
		return newError(ErrPeriodUnknown, "period", period)
	}
	return nil
}

func removeCategory(list []Category, id string) []Category {
	out := make([]Category, 0, len(list))
	for _, c := range list {
		if c.ID != id {
			out = append(out, c)
		}
	}
	return out
}

func insertCategory(list []Category, at int, c Category) []Category {
	list = append(list, Category{})
	copy(list[at+1:], list[at:])
	list[at] = c
	return list
}

func removeEntry(list []Entry, id string) []Entry {
	out := make([]Entry, 0, len(list))
	for _, e := range list {
		if e.ID != id {
			out = append(out, e)
		}
	}
	return out
}

func insertEntry(list []Entry, at int, e Entry) []Entry {
	list = append(list, Entry{})
	copy(list[at+1:], list[at:])
	list[at] = e
	return list
}
