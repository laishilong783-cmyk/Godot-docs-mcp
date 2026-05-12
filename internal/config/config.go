package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds all configuration for godot-docs-mcp.
type Config struct {
	DocsPath        string
	DBPath          string
	Version         string
	MaxResults      int
	MaxContentChars int
}

// Default values.
const (
	DefaultVersion         = "4.4"
	DefaultMaxResults      = 10
	DefaultMaxContentChars = 20000
)

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	docsPath := os.Getenv("GODOT_DOCS_PATH")
	if docsPath == "" {
		return nil, fmt.Errorf("GODOT_DOCS_PATH is not set")
	}

	docsPath, err := filepath.Abs(docsPath)
	if err != nil {
		return nil, fmt.Errorf("invalid GODOT_DOCS_PATH: %w", err)
	}

	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("GODOT_DOCS_PATH does not exist: %s", docsPath)
	}

	dbPath := os.Getenv("GODOT_DOCS_DB")
	if dbPath == "" {
		dbPath = filepath.Join(docsPath, "..", "godot_docs.db")
		dbPath, _ = filepath.Abs(dbPath)
	}

	version := os.Getenv("GODOT_DOCS_VERSION")
	if version == "" {
		version = DefaultVersion
	}

	maxResults := DefaultMaxResults
	if v := os.Getenv("GODOT_DOCS_MAX_RESULTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxResults = n
		}
	}

	maxContentChars := DefaultMaxContentChars
	if v := os.Getenv("GODOT_DOCS_MAX_CONTENT_CHARS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxContentChars = n
		}
	}

	return &Config{
		DocsPath:        docsPath,
		DBPath:          dbPath,
		Version:         version,
		MaxResults:      maxResults,
		MaxContentChars: maxContentChars,
	}, nil
}

// ValidateDocsPath checks if the docs path looks like a Godot docs repository.
func (c *Config) ValidateDocsPath() error {
	required := []string{"conf.py", "classes", "tutorials"}
	for _, name := range required {
		path := filepath.Join(c.DocsPath, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("GODOT_DOCS_PATH does not look like a Godot docs repo: missing %s", name)
		}
	}
	return nil
}
