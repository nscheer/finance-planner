package planner

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

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

// SampleResult reports what LoadSampleData created.
type SampleResult struct {
	State      State `json:"state"`
	Categories int   `json:"categories"`
	Entries    int   `json:"entries"`
}

// ImportResult is returned by ImportData: the new state plus a report of
// what the import did, shown to the user in a notification.
type ImportResult struct {
	State            State `json:"state"`
	CategoriesAdded  int   `json:"categoriesAdded"`
	CategoriesReused int   `json:"categoriesReused"`
	EntriesAdded     int   `json:"entriesAdded"`
	EntriesSkipped   int   `json:"entriesSkipped"`
	// BackupPath is the backup written right before the import; restoring
	// it undoes the import.
	BackupPath string `json:"backupPath"`
}

// Service is the Wails service used by the frontend. Every mutating method
// persists the data immediately and returns the new State, so the frontend
// simply replaces its state with the result.
type Service struct {
	mu   sync.Mutex
	path string
	data Data
	// lastBackup is when the last automatic backup was written.
	lastBackup time.Time
	// now is replaceable in tests.
	now func() time.Time
}

// NewService creates a service that persists to the given file. The file is
// loaded (or created empty) on construction.
func NewService(path string) (*Service, error) {
	d, err := LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", path, err)
	}
	s := &Service{path: path, data: d, now: time.Now}
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
// Before the change an automatic backup is written (rate limited).
func (s *Service) mutate(fn func() error) (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.backup(false); err != nil {
		return s.state(), err
	}
	if err := fn(); err != nil {
		return s.state(), err
	}
	if err := s.save(); err != nil {
		return s.state(), err
	}
	return s.state(), nil
}

// backup copies the data file into the backup folder. Ordinary edits only
// create a backup when the last one is older than backupInterval; force
// always creates one (used before imports and restores). Must be called
// with the lock held. Returns the backup path ("" when none was written).
func (s *Service) backup(force bool) (string, error) {
	now := s.now()
	if !force && now.Sub(s.lastBackup) < backupInterval {
		return "", nil
	}
	path, err := WriteBackup(s.path, now)
	if err != nil {
		return "", fmt.Errorf("writing backup: %w", err)
	}
	if path != "" {
		s.lastBackup = now
	}
	return path, nil
}

// ListBackups returns the available backups, newest first.
func (s *Service) ListBackups() ([]BackupInfo, error) {
	return ListBackups(s.path)
}

// RestoreBackup replaces the current data with a backup. The current data
// is backed up first, so a restore can be undone as well.
func (s *Service) RestoreBackup(path string) (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !isBackupPath(s.path, path) {
		return s.state(), newError(ErrBackupInvalidPath)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return s.state(), err
	}
	restored, err := Decode(raw)
	if err != nil {
		return s.state(), newError(ErrImportInvalidFile, "file", filepath.Base(path), "detail", err.Error())
	}
	if _, err := s.backup(true); err != nil {
		return s.state(), err
	}
	// Settings are preferences of this installation, not of the backup
	// (same rule as a replace import).
	restored.Settings = s.data.Settings
	s.data = restored
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
		s.data.Categories = removeCategory(s.data.Categories, id)
		s.data.Categories = insertCategory(s.data.Categories, s.categoryInsertPos(moved.Kind, toIndex), moved)
		return nil
	})
}

// ---- entries --------------------------------------------------------------

// EntryInput holds the editable fields of an entry as entered in the dialog.
type EntryInput struct {
	CategoryID  string `json:"categoryId"`
	Name        string `json:"name"`
	AmountCents int64  `json:"amountCents"`
	Period      Period `json:"period"`
	DueMonth    int    `json:"dueMonth"`
	Paused      bool   `json:"paused"`
	Notes       string `json:"notes"`
}

// AddEntry appends a new entry to a category.
func (s *Service) AddEntry(in EntryInput) (State, error) {
	return s.mutate(func() error {
		if err := s.validateEntry(&in); err != nil {
			return err
		}
		e := Entry{ID: newID()}
		in.applyTo(&e)
		s.data.Entries = append(s.data.Entries, e)
		return nil
	})
}

// UpdateEntry changes all editable fields of an entry. When the category
// changes, the entry is appended at the end of the new category.
func (s *Service) UpdateEntry(id string, in EntryInput) (State, error) {
	return s.mutate(func() error {
		e := s.data.Entry(id)
		if e == nil {
			return newError(ErrEntryNotFound)
		}
		if err := s.validateEntry(&in); err != nil {
			return err
		}
		moved := e.CategoryID != in.CategoryID
		in.applyTo(e)
		if moved {
			copied := *e
			s.data.Entries = append(removeEntry(s.data.Entries, id), copied)
		}
		return nil
	})
}

// SetEntryPaused excludes an entry from (or includes it again in) all totals.
func (s *Service) SetEntryPaused(id string, paused bool) (State, error) {
	return s.mutate(func() error {
		e := s.data.Entry(id)
		if e == nil {
			return newError(ErrEntryNotFound)
		}
		e.Paused = paused
		return nil
	})
}

// RestoreEntry re-inserts a deleted entry with its original id at the given
// position of its category ("undo" of DeleteEntry).
func (s *Service) RestoreEntry(entry Entry, index int) (State, error) {
	return s.mutate(func() error {
		if entry.ID == "" || s.data.Entry(entry.ID) != nil {
			return newError(ErrEntryIDExists)
		}
		in := EntryInput{
			CategoryID: entry.CategoryID, Name: entry.Name, AmountCents: entry.AmountCents,
			Period: entry.Period, DueMonth: entry.DueMonth, Paused: entry.Paused, Notes: entry.Notes,
		}
		if err := s.validateEntry(&in); err != nil {
			return err
		}
		e := Entry{ID: entry.ID}
		in.applyTo(&e)
		s.data.Entries = insertEntry(s.data.Entries, s.entryInsertPos(e.CategoryID, index), e)
		return nil
	})
}

// RestoreCategory re-inserts a deleted (empty) category with its original
// id at the given position among the categories of its kind.
func (s *Service) RestoreCategory(category Category, index int) (State, error) {
	return s.mutate(func() error {
		category.Name = strings.TrimSpace(category.Name)
		if category.ID == "" || s.data.Category(category.ID) != nil {
			return newError(ErrCategoryIDExists)
		}
		if !category.Kind.Valid() {
			return newError(ErrKindUnknown, "kind", category.Kind)
		}
		if category.Name == "" {
			return newError(ErrCategoryNameEmpty)
		}
		if s.findCategoryByName(category.Kind, category.Name) != nil {
			return newError(ErrCategoryExists, "kind", category.Kind, "name", category.Name)
		}
		s.data.Categories = insertCategory(s.data.Categories, s.categoryInsertPos(category.Kind, index), category)
		return nil
	})
}

// ---- bulk operations on a selection of entries ----------------------------

// EntryAt is a deleted entry together with its former position inside its
// category, so that a bulk delete can be undone.
type EntryAt struct {
	Entry Entry `json:"entry"`
	Index int   `json:"index"`
}

// lookupEntries resolves ids and fails on the first unknown one.
func (s *Service) lookupEntries(ids []string) ([]*Entry, error) {
	out := make([]*Entry, 0, len(ids))
	for _, id := range ids {
		e := s.data.Entry(id)
		if e == nil {
			return nil, newError(ErrEntryNotFound)
		}
		out = append(out, e)
	}
	return out, nil
}

// MoveEntries appends the given entries, in the given order, to the end of
// the target category. All entries must belong to the target's kind.
func (s *Service) MoveEntries(ids []string, targetCategoryID string) (State, error) {
	return s.mutate(func() error {
		to := s.data.Category(targetCategoryID)
		if to == nil {
			return newError(ErrCategoryNotFound)
		}
		entries, err := s.lookupEntries(ids)
		if err != nil {
			return err
		}
		moved := make([]Entry, 0, len(entries))
		for _, e := range entries {
			if from := s.data.Category(e.CategoryID); from != nil && from.Kind != to.Kind {
				return newError(ErrEntryKindMismatch, "from", from.Kind, "to", to.Kind)
			}
			copied := *e
			copied.CategoryID = targetCategoryID
			moved = append(moved, copied)
		}
		for _, id := range ids {
			s.data.Entries = removeEntry(s.data.Entries, id)
		}
		for _, e := range moved {
			s.data.Entries = insertEntry(s.data.Entries, s.entryInsertPos(targetCategoryID, len(s.data.Entries)), e)
		}
		return nil
	})
}

// SetEntriesPaused pauses or resumes several entries at once.
func (s *Service) SetEntriesPaused(ids []string, paused bool) (State, error) {
	return s.mutate(func() error {
		entries, err := s.lookupEntries(ids)
		if err != nil {
			return err
		}
		for _, e := range entries {
			e.Paused = paused
		}
		return nil
	})
}

// DeleteEntries removes several entries at once.
func (s *Service) DeleteEntries(ids []string) (State, error) {
	return s.mutate(func() error {
		if _, err := s.lookupEntries(ids); err != nil {
			return err
		}
		for _, id := range ids {
			s.data.Entries = removeEntry(s.data.Entries, id)
		}
		return nil
	})
}

// RestoreEntries re-inserts deleted entries at their former positions
// ("undo" of DeleteEntries). Items are applied in ascending index order per
// category, so the original order is reproduced.
func (s *Service) RestoreEntries(items []EntryAt) (State, error) {
	return s.mutate(func() error {
		sorted := append([]EntryAt(nil), items...)
		sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Index < sorted[j].Index })
		for _, it := range sorted {
			entry := it.Entry
			if entry.ID == "" || s.data.Entry(entry.ID) != nil {
				return newError(ErrEntryIDExists)
			}
			in := EntryInput{
				CategoryID: entry.CategoryID, Name: entry.Name, AmountCents: entry.AmountCents,
				Period: entry.Period, DueMonth: entry.DueMonth, Paused: entry.Paused, Notes: entry.Notes,
			}
			if err := s.validateEntry(&in); err != nil {
				return err
			}
			e := Entry{ID: entry.ID}
			in.applyTo(&e)
			s.data.Entries = insertEntry(s.data.Entries, s.entryInsertPos(e.CategoryID, it.Index), e)
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
		s.data.Entries = removeEntry(s.data.Entries, id)
		s.data.Entries = insertEntry(s.data.Entries, s.entryInsertPos(targetCategoryID, toIndex), moved)
		return nil
	})
}

// ---- settings -------------------------------------------------------------

// SetTheme stores the chosen color scheme ("light", "dark" or "system").
func (s *Service) SetTheme(theme string) (State, error) {
	return s.mutate(func() error {
		if theme == "" || !ValidTheme(theme) {
			return newError(ErrThemeInvalid, "theme", theme)
		}
		s.data.Settings.Theme = theme
		return nil
	})
}

// SetSavingsGoal stores the monthly savings goal.
func (s *Service) SetSavingsGoal(cents int64) (State, error) {
	return s.mutate(func() error {
		if cents < 0 {
			return newError(ErrSavingsGoalNegative)
		}
		s.data.Settings.SavingsGoalCents = cents
		return nil
	})
}

// SetWindow remembers the window geometry. It is called often while the
// window is resized, so nothing is written when the geometry is unchanged.
func (s *Service) SetWindow(w WindowGeometry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.Settings.Window == w {
		return nil
	}
	s.data.Settings.Window = w
	return s.save()
}

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
		backupPath, err := s.backup(true)
		if err != nil {
			return err
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
		result.BackupPath = backupPath
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

// ExportCSVTo writes all entries as CSV to the given path.
func (s *Service) ExportCSVTo(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var buf bytes.Buffer
	if err := WriteCSV(&buf, s.data, s.data.Settings.Language); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// LoadSampleData fills an empty planner with example data in the given
// language. It refuses to run when there already are categories or entries.
func (s *Service) LoadSampleData(lang string) (SampleResult, error) {
	var result SampleResult
	st, err := s.mutate(func() error {
		if len(s.data.Categories) > 0 || len(s.data.Entries) > 0 {
			return newError(ErrSampleNotEmpty)
		}
		sample := SampleData(lang)
		sample.Settings = s.data.Settings
		s.data = sample
		result.Categories = len(sample.Categories)
		result.Entries = len(sample.Entries)
		return nil
	})
	result.State = st
	return result, err
}

// ---- dialogs (need a running Wails application) ---------------------------

// ExportCSV asks the user for a target file and writes the CSV to it.
// It returns the chosen path, or "" if the user cancelled.
func (s *Service) ExportCSV() (string, error) {
	path, err := application.Get().Dialog.SaveFile().
		SetMessage("Export CSV").
		SetFilename("finance-planner-export.csv").
		AddFilter("CSV files", "*.csv").
		PromptForSingleSelection()
	if err != nil || path == "" {
		return "", err
	}
	if filepath.Ext(path) == "" {
		path += ".csv"
	}
	return path, s.ExportCSVTo(path)
}

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

// validateEntry checks and normalises the input in place.
func (s *Service) validateEntry(in *EntryInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Notes = strings.TrimSpace(in.Notes)
	if s.data.Category(in.CategoryID) == nil {
		return newError(ErrEntryCategoryRequired)
	}
	if in.Name == "" {
		return newError(ErrEntryNameEmpty)
	}
	if in.AmountCents <= 0 {
		return newError(ErrEntryAmountPositive)
	}
	if !in.Period.Valid() {
		return newError(ErrPeriodUnknown, "period", in.Period)
	}
	if in.DueMonth < 0 || in.DueMonth > 12 {
		return newError(ErrEntryDueMonthInvalid, "month", in.DueMonth)
	}
	if in.Period == PeriodMonthly {
		in.DueMonth = 0 // monthly entries have no due month
	}
	return nil
}

// applyTo copies the input fields onto an entry (id is left untouched).
func (in EntryInput) applyTo(e *Entry) {
	e.CategoryID = in.CategoryID
	e.Name = in.Name
	e.AmountCents = in.AmountCents
	e.Period = in.Period
	e.DueMonth = in.DueMonth
	e.Paused = in.Paused
	e.Notes = in.Notes
}

// categoryInsertPos returns the global slice position at which a category
// has to be inserted to become the index-th category of its kind; past the
// end means "after the last one of that kind".
func (s *Service) categoryInsertPos(kind Kind, index int) int {
	insertAt := len(s.data.Categories)
	seen := 0
	for i, other := range s.data.Categories {
		if other.Kind != kind {
			continue
		}
		if seen == index {
			return i
		}
		seen++
		insertAt = i + 1
	}
	return insertAt
}

// entryInsertPos is the entry counterpart of categoryInsertPos.
func (s *Service) entryInsertPos(categoryID string, index int) int {
	insertAt := len(s.data.Entries)
	seen := 0
	for i, other := range s.data.Entries {
		if other.CategoryID != categoryID {
			continue
		}
		if seen == index {
			return i
		}
		seen++
		insertAt = i + 1
	}
	return insertAt
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
