package planner

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultDataFileName is the name of the data file next to the binary.
const DefaultDataFileName = "data.json"

// DefaultDataPath returns the path of data.json next to the running binary.
func DefaultDataPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return filepath.Join(filepath.Dir(exe), DefaultDataFileName), nil
}

// LoadFile reads and validates a data file. A missing file yields an empty,
// current-version data set so that a fresh installation works without setup.
func LoadFile(path string) (Data, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewData(), nil
	}
	if err != nil {
		return Data{}, err
	}
	return Decode(raw)
}

// Decode parses JSON bytes, migrates older versions and validates the result.
func Decode(raw []byte) (Data, error) {
	var probe struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return Data{}, fmt.Errorf("invalid JSON: %w", err)
	}
	if probe.Version > CurrentVersion {
		return Data{}, fmt.Errorf("data version %d is newer than supported version %d", probe.Version, CurrentVersion)
	}
	raw, err := migrate(raw, probe.Version)
	if err != nil {
		return Data{}, err
	}
	var d Data
	if err := json.Unmarshal(raw, &d); err != nil {
		return Data{}, fmt.Errorf("invalid data file: %w", err)
	}
	if d.Categories == nil {
		d.Categories = []Category{}
	}
	if d.Entries == nil {
		d.Entries = []Entry{}
	}
	if d.Settings.Language == "" {
		d.Settings.Language = DefaultLanguage
	}
	d.Version = CurrentVersion
	if err := d.Validate(); err != nil {
		return Data{}, err
	}
	return d, nil
}

// migrate converts the raw JSON of an older version step by step to the
// current version. Each case upgrades exactly one version; add new cases at
// the bottom when CurrentVersion is bumped.
func migrate(raw []byte, version int) ([]byte, error) {
	for version < CurrentVersion {
		switch version {
		case 0:
			// Version 0 means "no version field". The structure equals
			// version 1, so only the version number needs to be set.
			var m map[string]json.RawMessage
			if err := json.Unmarshal(raw, &m); err != nil {
				return nil, fmt.Errorf("invalid JSON: %w", err)
			}
			m["version"] = json.RawMessage("1")
			var err error
			if raw, err = json.Marshal(m); err != nil {
				return nil, err
			}
			version = 1
		case 1:
			// Version 2 added the "settings" object.
			var m map[string]json.RawMessage
			if err := json.Unmarshal(raw, &m); err != nil {
				return nil, fmt.Errorf("invalid JSON: %w", err)
			}
			if _, ok := m["settings"]; !ok {
				m["settings"] = json.RawMessage(`{"language":"` + DefaultLanguage + `"}`)
			}
			m["version"] = json.RawMessage("2")
			var err error
			if raw, err = json.Marshal(m); err != nil {
				return nil, err
			}
			version = 2
		default:
			return nil, fmt.Errorf("no migration from version %d", version)
		}
	}
	return raw, nil
}

// Encode serialises data as indented JSON.
func Encode(d Data) ([]byte, error) {
	d.Version = CurrentVersion
	return json.MarshalIndent(d, "", "  ")
}

// SaveFile writes the data atomically (write to temp file, then rename) so
// that a crash during writing never leaves a truncated data.json behind.
func SaveFile(path string, d Data) error {
	raw, err := Encode(d)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".data-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
