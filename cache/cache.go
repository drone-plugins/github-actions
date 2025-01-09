package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/lockedfile"
	"golang.org/x/exp/slog"
)

const (
	completionMarkerFile = ".done"
)

func Add(key string, addItem func() error) error {
	if err := os.MkdirAll(key, 0700); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create directory %s", key))
	}

	lockFilepath := filepath.Join(key, ".started")
	slog.Debug("taking lock", "key", lockFilepath)
	lock, err := lockedfile.Create(lockFilepath)
	slog.Debug("took lock", "key", lockFilepath)

	if err != nil {
		return errors.Wrap(err, "failed to take file lock")
	}
	defer func() {
		if err := lock.Close(); err != nil {
			slog.Error("failed to release lock", "key", lockFilepath, "error", err)
		}
		slog.Debug("released lock", "key", lockFilepath)
	}()
	// If data is already present, return
	if _, err := os.Stat(filepath.Join(key, completionMarkerFile)); err == nil {
		return nil
	}

	if err := addItem(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to add item: %s to cache", key))
	}

	integrityFpath := filepath.Join(key, completionMarkerFile)
	f, err := os.Create(integrityFpath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create integrity file: %s", integrityFpath))
	}
	f.Close()

	return nil
}

// GetKeyName generate unique file path inside cache directory
// based on name provided
func GetKeyName(name string) string {
	return filepath.Join(getCacheDir(), sha(name))
}

func getCacheDir() string {
	dir, _ := os.UserHomeDir()
	return filepath.Join(dir, ".cache")
}

func sha(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
