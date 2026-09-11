package tests_test

import (
	"fmt"
	"os"
	"path/filepath"
)

// locateTgz returns the absolute path of the single *.tgz file in dir.
// It returns an error when dir does not exist, contains no *.tgz file, or
// contains more than one.
func locateTgz(dir string) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("locateTgz: directory %q does not exist", dir)
		}
		return "", fmt.Errorf("locateTgz: stat %q: %w", dir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("locateTgz: %q is not a directory", dir)
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.tgz"))
	if err != nil {
		return "", fmt.Errorf("locateTgz: glob %q: %w", filepath.Join(dir, "*.tgz"), err)
	}
	switch len(matches) {
	case 1:
		absolute, err := filepath.Abs(matches[0])
		if err != nil {
			return "", fmt.Errorf("locateTgz: resolve %q: %w", matches[0], err)
		}
		return absolute, nil
	case 0:
		return "", fmt.Errorf("locateTgz: no *.tgz file found in %q", dir)
	default:
		return "", fmt.Errorf("locateTgz: expected exactly one *.tgz file in %q, found %d", dir, len(matches))
	}
}
