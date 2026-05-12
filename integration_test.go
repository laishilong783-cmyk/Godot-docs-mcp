package main

import (
	"path/filepath"
	"testing"

	"github.com/godot-docs-mcp/godot-docs-mcp/internal/docs"
	"github.com/godot-docs-mcp/godot-docs-mcp/internal/index"
)

func TestFixtureIndexAndSearch(t *testing.T) {
	fixturePath := filepath.Join("testdata", "fixtures")
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := index.Open(dbPath)
	if err != nil {
		t.Fatalf("Open db: %v", err)
	}
	defer db.Close()

	version := "4.4-test"

	files, err := docs.Scan(fixturePath)
	if err != nil {
		t.Fatalf("Scan fixtures: %v", err)
	}

	for _, f := range files {
		parsed, err := docs.ParseFile(fixturePath, f)
		if err != nil {
			t.Logf("ParseFile %s: %v", f.Path, err)
			continue
		}
		_, err = db.InsertDocument(version, parsed.Path, parsed.Title, parsed.Section, parsed.Content)
		if err != nil {
			t.Fatalf("InsertDocument: %v", err)
		}
	}

	if err := db.SyncFTS(); err != nil {
		t.Fatalf("SyncFTS: %v", err)
	}

	// Search for CharacterBody2D.
	results, err := db.Search(version, "CharacterBody2D", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Expected search results for CharacterBody2D")
	}

	found := false
	for _, r := range results {
		if r.Title == "CharacterBody2D" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'CharacterBody2D' in results, got %+v", results)
	}
}
