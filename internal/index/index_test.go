package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAndMigrate(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database file was not created")
	}
}

func TestInsertAndSearch(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	version := "4.4"

	// Insert documents.
	_, err = db.InsertDocument(version, "classes/class_node.rst", "Node", "classes", "Node is the base class.")
	if err != nil {
		t.Fatalf("InsertDocument failed: %v", err)
	}
	_, err = db.InsertDocument(version, "tutorials/physics/body.rst", "Physics Body", "tutorials", "Using physics bodies.")
	if err != nil {
		t.Fatalf("InsertDocument failed: %v", err)
	}

	if err := db.SyncFTS(); err != nil {
		t.Fatalf("SyncFTS failed: %v", err)
	}

	// Search.
	results, err := db.Search(version, "Node", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Expected search results for 'Node'")
	}

	found := false
	for _, r := range results {
		if r.Title == "Node" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'Node' in search results, got %+v", results)
	}
}

func TestInsertSymbol(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	version := "4.4"
	err = db.InsertSymbol(version, "class", "Node", "", "", "", "Base node class.", "classes/class_node.rst", 0, 0)
	if err != nil {
		t.Fatalf("InsertSymbol class failed: %v", err)
	}
	err = db.InsertSymbol(version, "method", "Node", "add_child", "void add_child(Node node)", "void", "Adds a child.", "classes/class_node.rst", 10, 20)
	if err != nil {
		t.Fatalf("InsertSymbol method failed: %v", err)
	}

	// Get class.
	info, err := db.GetClass(version, "Node")
	if err != nil {
		t.Fatalf("GetClass failed: %v", err)
	}
	if info["class_name"] != "Node" {
		t.Errorf("GetClass class_name = %v, want Node", info["class_name"])
	}

	// Get method.
	method, err := db.GetMethod(version, "Node", "add_child")
	if err != nil {
		t.Fatalf("GetMethod failed: %v", err)
	}
	if method["method_name"] != "add_child" {
		t.Errorf("GetMethod method_name = %v, want add_child", method["method_name"])
	}
}

func TestGetPage(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	version := "4.4"
	_, err = db.InsertDocument(version, "tutorials/test.rst", "Test", "tutorials", "Test content here.")
	if err != nil {
		t.Fatalf("InsertDocument failed: %v", err)
	}

	title, section, content, err := db.GetPage(version, "tutorials/test.rst")
	if err != nil {
		t.Fatalf("GetPage failed: %v", err)
	}
	if title != "Test" {
		t.Errorf("title = %q, want Test", title)
	}
	if section != "tutorials" {
		t.Errorf("section = %q, want tutorials", section)
	}
	if content != "Test content here." {
		t.Errorf("content = %q, want Test content here.", content)
	}
}
