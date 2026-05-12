package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/godot-docs-mcp/godot-docs-mcp/internal/config"
	"github.com/godot-docs-mcp/godot-docs-mcp/internal/docs"
	"github.com/godot-docs-mcp/godot-docs-mcp/internal/index"
)

func main() {
	var (
		docsPath = flag.String("docs-path", "", "Path to local godot-docs repo")
		version  = flag.String("version", config.DefaultVersion, "Godot docs version")
		dbPath   = flag.String("db", "", "SQLite database path")
	)
	flag.Parse()

	if *docsPath == "" {
		*docsPath = os.Getenv("GODOT_DOCS_PATH")
	}
	if *docsPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --docs-path or GODOT_DOCS_PATH is required")
		flag.Usage()
		os.Exit(1)
	}

	if *dbPath == "" {
		*dbPath = os.Getenv("GODOT_DOCS_DB")
		if *dbPath == "" {
			*dbPath = filepath.Join(*docsPath, "..", fmt.Sprintf("godot_docs_%s.db", strings.ReplaceAll(*version, ".", "_")))
			*dbPath, _ = filepath.Abs(*dbPath)
		}
	}

	cfg := &config.Config{
		DocsPath: *docsPath,
		DBPath:   *dbPath,
		Version:  *version,
	}

	if err := cfg.ValidateDocsPath(); err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	fmt.Printf("Godot Docs Index Build\n")
	fmt.Printf("Version:   %s\n", *version)
	fmt.Printf("Docs path: %s\n", cfg.DocsPath)
	fmt.Printf("Database:  %s\n", cfg.DBPath)
	fmt.Println()

	start := time.Now()

	db, err := index.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Open database: %v", err)
	}
	defer db.Close()

	// Clear existing data for this version.
	if err := db.ClearVersion(*version); err != nil {
		log.Fatalf("Clear version: %v", err)
	}

	// Scan documents.
	files, err := docs.Scan(cfg.DocsPath)
	if err != nil {
		log.Fatalf("Scan docs: %v", err)
	}

	docCount := 0
	symbolCount := 0

	for _, f := range files {
		parsed, err := docs.ParseFile(cfg.DocsPath, f)
		if err != nil {
			continue
		}

		_, err = db.InsertDocument(*version, parsed.Path, parsed.Title, parsed.Section, parsed.Content)
		if err != nil {
			log.Printf("Insert document %s: %v", parsed.Path, err)
			continue
		}
		docCount++

		// Parse class symbols from classes/ directory.
		if f.Section == "classes" && strings.HasPrefix(filepath.Base(f.Path), "class_") {
			_, symbols, err := docs.ParseClassDoc(cfg.DocsPath, f)
			if err != nil {
				continue
			}
			for _, sym := range symbols {
				if err := db.InsertSymbol(*version, string(sym.Kind), sym.ClassName, sym.MemberName,
					sym.Signature, sym.ReturnType, sym.Description, sym.Path, sym.LineStart, sym.LineEnd); err != nil {
					log.Printf("Insert symbol %s.%s: %v", sym.ClassName, sym.MemberName, err)
					continue
				}
				symbolCount++
			}
		}
	}

	// Rebuild FTS index.
	fmt.Println("Rebuilding FTS index...")
	if err := db.SyncFTS(); err != nil {
		log.Fatalf("Sync FTS: %v", err)
	}

	// Stats.
	stats, err := db.Stats(*version)
	if err != nil {
		log.Fatalf("Stats: %v", err)
	}

	elapsed := time.Since(start)

	fmt.Println()
	fmt.Println("Godot Docs Index Build Complete")
	fmt.Printf("Version:          %s\n", *version)
	fmt.Printf("Docs path:        %s\n", cfg.DocsPath)
	fmt.Printf("Database:         %s\n", cfg.DBPath)
	fmt.Printf("Documents indexed: %d\n", docCount)
	for kind, count := range stats {
		if kind != "documents" {
			fmt.Printf("%s indexed: %d\n", strings.Title(kind)+"s", count)
		}
	}
	fmt.Printf("Total symbols:    %d\n", symbolCount)
	fmt.Printf("Elapsed:          %.1fs\n", elapsed.Seconds())
}
