package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := `# Metro API
METRO_API_URL=https://api.example/v1/
export METRO_API_TOKEN="se cret"
PORT=8081 # behind nginx
CAIS_ALREADY=from-file
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CAIS_ALREADY", "from-env")
	for _, k := range []string{"METRO_API_URL", "METRO_API_TOKEN", "PORT"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}

	if err := LoadEnvFile(path); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"METRO_API_URL":   "https://api.example/v1/",
		"METRO_API_TOKEN": "se cret",
		"PORT":            "8081",
		"CAIS_ALREADY":    "from-env", // the real environment wins
	}
	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestLoadEnvFileMissingIsFine(t *testing.T) {
	if err := LoadEnvFile(filepath.Join(t.TempDir(), "nope")); err != nil {
		t.Errorf("a missing file should be ignored, got %v", err)
	}
}

func TestLoadEnvFileRejectsGarbage(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(path, []byte("just words\n"), 0o600)
	if err := LoadEnvFile(path); err == nil {
		t.Error("a line without = should be an error")
	}
}
