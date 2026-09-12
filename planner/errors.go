package planner

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Error is a user-facing error with a stable code, so that the frontend can
// show it in the user's language. Params carry the values to interpolate.
//
// The Error() text is an English fallback for logs and tests; the frontend
// never shows it when it knows the code.
type Error struct {
	Code   string         `json:"code"`
	Params map[string]any `json:"params,omitempty"`
}

// Error codes returned by the service. The frontend has a translation for
// each of them under "errors.<code>".
const (
	ErrKindUnknown           = "kind.unknown"
	ErrPeriodUnknown         = "period.unknown"
	ErrCategoryNotFound      = "category.notFound"
	ErrCategoryNameEmpty     = "category.nameEmpty"
	ErrCategoryExists        = "category.exists"
	ErrCategoryInUse         = "category.inUse"
	ErrEntryNotFound         = "entry.notFound"
	ErrEntryCategoryRequired = "entry.categoryRequired"
	ErrEntryNameEmpty        = "entry.nameEmpty"
	ErrEntryAmountPositive   = "entry.amountPositive"
	ErrEntryDueMonthInvalid  = "entry.dueMonthInvalid"
	ErrSavingsGoalNegative   = "settings.savingsGoalNegative"
	ErrEntryKindMismatch     = "entry.kindMismatch"
	ErrImportModeUnknown     = "import.modeUnknown"
	ErrImportInvalidFile     = "import.invalidFile"
	ErrLanguageInvalid       = "language.invalid"
	ErrEntryIDExists         = "entry.idExists"
	ErrCategoryIDExists      = "category.idExists"
	ErrSampleNotEmpty        = "sample.notEmpty"
	ErrBackupInvalidPath     = "backup.invalidPath"
)

// newError creates a coded error. params are alternating key/value pairs.
func newError(code string, params ...any) *Error {
	e := &Error{Code: code}
	if len(params) > 0 {
		e.Params = map[string]any{}
		for i := 0; i+1 < len(params); i += 2 {
			e.Params[fmt.Sprint(params[i])] = params[i+1]
		}
	}
	return e
}

func (e *Error) Error() string {
	if len(e.Params) == 0 {
		return e.Code
	}
	keys := make([]string, 0, len(e.Params))
	for k := range e.Params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, e.Params[k]))
	}
	return e.Code + " (" + strings.Join(parts, ", ") + ")"
}

// MarshalError is used as ServiceOptions.MarshalError: coded errors are sent
// to the frontend as JSON (available as `error.cause` there). For any other
// error nil is returned, which makes Wails fall back to its default handling.
func MarshalError(err error) []byte {
	var e *Error
	if !errors.As(err, &e) {
		return nil
	}
	raw, marshalErr := json.Marshal(e)
	if marshalErr != nil {
		return nil
	}
	return raw
}
