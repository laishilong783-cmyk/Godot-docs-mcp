package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("GODOT_DOCS_PATH", tmpDir)
	defer os.Unsetenv("GODOT_DOCS_PATH")

	// Create a fake docs structure.
	os.WriteFile(filepath.Join(tmpDir, "conf.py"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "classes"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "tutorials"), 0755)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.DocsPath != tmpDir {
		t.Errorf("DocsPath = %s, want %s", cfg.DocsPath, tmpDir)
	}
	if cfg.Version != DefaultVersion {
		t.Errorf("Version = %s, want %s", cfg.Version, DefaultVersion)
	}
}

func TestLoadMissingPath(t *testing.T) {
	os.Unsetenv("GODOT_DOCS_PATH")
	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing GODOT_DOCS_PATH")
	}
}

func TestValidateDocsPath(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{DocsPath: tmpDir}

	if err := cfg.ValidateDocsPath(); err == nil {
		t.Fatal("Expected validation error for empty dir")
	}

	os.WriteFile(filepath.Join(tmpDir, "conf.py"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "classes"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "tutorials"), 0755)

	if err := cfg.ValidateDocsPath(); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
}
