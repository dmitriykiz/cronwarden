package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/cronwarden/internal/config"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeConfig: %v", err)
	}
	return p
}

func TestLoad_Valid(t *testing.T) {
	path := writeConfig(t, `{
		"db_path": "/tmp/test.db",
		"jobs": [
			{"name": "backup", "schedule": "@daily", "command": "/usr/bin/backup.sh", "webhook_url": "http://example.com/hook"}
		]
	}`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("db_path: got %q, want %q", cfg.DBPath, "/tmp/test.db")
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("jobs count: got %d, want 1", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("job name: got %q", cfg.Jobs[0].Name)
	}
}

func TestLoad_DefaultDBPath(t *testing.T) {
	path := writeConfig(t, `{"jobs":[{"name":"ping","schedule":"* * * * *","command":"echo ok"}]}`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBPath != "cronwarden.db" {
		t.Errorf("default db_path: got %q, want %q", cfg.DBPath, "cronwarden.db")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeConfig(t, `{"db_path": "x.db", "jobs": []}`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty jobs")
	}
}

func TestLoad_JobMissingCommand(t *testing.T) {
	path := writeConfig(t, `{"jobs":[{"name":"bad","schedule":"@hourly"}]}`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing command")
	}
}
