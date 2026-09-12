package planner

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// BackupDirName is the folder next to data.json that holds backups.
	BackupDirName = "backups"
	// MaxBackups is the number of backups kept; older ones are pruned.
	MaxBackups = 20
	// backupInterval limits how often ordinary edits create a backup.
	// Imports and restores always create one (see Service.backup).
	backupInterval = 10 * time.Minute
)

// BackupInfo describes one backup file for the restore dialog.
type BackupInfo struct {
	Path       string    `json:"path"`
	Time       time.Time `json:"time"`
	Categories int       `json:"categories"`
	Entries    int       `json:"entries"`
}

// BackupDir returns the backup folder for a data file.
func BackupDir(dataPath string) string {
	return filepath.Join(filepath.Dir(dataPath), BackupDirName)
}

// WriteBackup copies the current data file into the backup folder and prunes
// old backups. It returns the backup path, or "" if there was nothing to
// back up (no data file yet).
func WriteBackup(dataPath string, now time.Time) (string, error) {
	raw, err := os.ReadFile(dataPath)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	dir := BackupDir(dataPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := "data-" + now.Format("20060102-150405.000") + ".json"
	target := filepath.Join(dir, name)
	if err := os.WriteFile(target, raw, 0o644); err != nil {
		return "", err
	}
	return target, pruneBackups(dir)
}

// pruneBackups keeps only the newest MaxBackups files.
func pruneBackups(dir string) error {
	infos, err := listBackupFiles(dir)
	if err != nil {
		return err
	}
	for i := MaxBackups; i < len(infos); i++ {
		if err := os.Remove(infos[i].Path); err != nil {
			return err
		}
	}
	return nil
}

// listBackupFiles returns the backups in the folder, newest first.
func listBackupFiles(dir string) ([]BackupInfo, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []BackupInfo
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "data-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		stamp := strings.TrimSuffix(strings.TrimPrefix(name, "data-"), ".json")
		t, err := time.ParseInLocation("20060102-150405.000", stamp, time.Local)
		if err != nil {
			continue
		}
		out = append(out, BackupInfo{Path: filepath.Join(dir, name), Time: t})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	return out, nil
}

// ListBackups returns the available backups (newest first) including the
// number of categories and entries each one contains.
func ListBackups(dataPath string) ([]BackupInfo, error) {
	infos, err := listBackupFiles(BackupDir(dataPath))
	if err != nil {
		return nil, err
	}
	for i := range infos {
		raw, err := os.ReadFile(infos[i].Path)
		if err != nil {
			continue
		}
		if d, err := Decode(raw); err == nil {
			infos[i].Categories = len(d.Categories)
			infos[i].Entries = len(d.Entries)
		}
	}
	if infos == nil {
		infos = []BackupInfo{}
	}
	return infos, nil
}

// isBackupPath reports whether path points to a file inside the backup
// folder of dataPath (so that RestoreBackup can't be used to read other files).
func isBackupPath(dataPath, path string) bool {
	dir, err := filepath.Abs(BackupDir(dataPath))
	if err != nil {
		return false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return filepath.Dir(abs) == dir && strings.HasSuffix(abs, ".json")
}
