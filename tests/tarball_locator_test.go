package tests_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocateTgz(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string) string
		wantErr bool
		missing bool
	}{
		{
			name: "returns the absolute path of the single tgz file",
			setup: func(t *testing.T, dir string) string {
				writeTestFile(t, filepath.Join(dir, "README.txt"))
				writeTestFile(t, filepath.Join(dir, "package.tgz"))
				return filepath.Join(dir, "package.tgz")
			},
		},
		{
			name: "errors when no tgz file is present",
			setup: func(t *testing.T, dir string) string {
				writeTestFile(t, filepath.Join(dir, "README.txt"))
				return ""
			},
			wantErr: true,
		},
		{
			name: "errors when more than one tgz file is present",
			setup: func(t *testing.T, dir string) string {
				writeTestFile(t, filepath.Join(dir, "first.tgz"))
				writeTestFile(t, filepath.Join(dir, "second.tgz"))
				return ""
			},
			wantErr: true,
		},
		{
			name: "errors when the directory does not exist",
			setup: func(t *testing.T, dir string) string {
				return ""
			},
			wantErr: true,
			missing: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			want := tt.setup(t, dir)
			if tt.missing {
				dir = filepath.Join(dir, "missing")
			}

			got, err := locateTgz(dir)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("locateTgz(%q) error = nil, want an error", dir)
				}
				return
			}
			if err != nil {
				t.Fatalf("locateTgz(%q) error = %v", dir, err)
			}
			if !filepath.IsAbs(got) {
				t.Fatalf("locateTgz(%q) = %q, want an absolute path", dir, got)
			}
			if got != want {
				t.Fatalf("locateTgz(%q) = %q, want %q", dir, got, want)
			}
		})
	}

	t.Run("is deterministic regardless of entry creation order", func(t *testing.T) {
		firstDir := t.TempDir()
		writeTestFile(t, filepath.Join(firstDir, "a.txt"))
		writeTestFile(t, filepath.Join(firstDir, "target.tgz"))
		writeTestFile(t, filepath.Join(firstDir, "z.txt"))

		secondDir := t.TempDir()
		writeTestFile(t, filepath.Join(secondDir, "z.txt"))
		writeTestFile(t, filepath.Join(secondDir, "target.tgz"))
		writeTestFile(t, filepath.Join(secondDir, "a.txt"))

		first, err := locateTgz(firstDir)
		if err != nil {
			t.Fatalf("locateTgz(%q) error = %v", firstDir, err)
		}
		second, err := locateTgz(secondDir)
		if err != nil {
			t.Fatalf("locateTgz(%q) error = %v", secondDir, err)
		}
		if first != filepath.Join(firstDir, "target.tgz") {
			t.Fatalf("first result = %q, want %q", first, filepath.Join(firstDir, "target.tgz"))
		}
		if second != filepath.Join(secondDir, "target.tgz") {
			t.Fatalf("second result = %q, want %q", second, filepath.Join(secondDir, "target.tgz"))
		}
	})
}

func writeTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
		t.Fatalf("write fixture %q: %v", path, err)
	}
}
