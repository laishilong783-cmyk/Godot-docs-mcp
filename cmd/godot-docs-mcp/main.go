package main

import (
	"fmt"
	"log"
	"os"

	"github.com/godot-docs-mcp/godot-docs-mcp/internal/config"
	"github.com/godot-docs-mcp/godot-docs-mcp/internal/index"
	mcpServer "github.com/godot-docs-mcp/godot-docs-mcp/internal/mcp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, `Error: %v

Please set the required environment variables:

  GODOT_DOCS_PATH    - Path to local clone of https://github.com/godotengine/godot-docs
  GODOT_DOCS_DB      - SQLite database path (optional, defaults to godot_docs.db next to docs)
  GODOT_DOCS_VERSION - Default Godot version (optional, defaults to %s)

Example:
  $env:GODOT_DOCS_PATH = "C:\Users\you\dev\godot-docs"
  $env:GODOT_DOCS_DB = "C:\Users\you\dev\godot_docs.db"
  $env:GODOT_DOCS_VERSION = "4.4"
`, err, config.DefaultVersion)
		os.Exit(1)
	}

	if err := cfg.ValidateDocsPath(); err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	// Check database exists.
	if _, err := os.Stat(cfg.DBPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, `Error: Index database not found at %s

Please run godot-docs-index first:

  godot-docs-index --docs-path "%s" --version %s --db "%s"
`, cfg.DBPath, cfg.DocsPath, cfg.Version, cfg.DBPath)
		os.Exit(1)
	}

	db, err := index.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Open database: %v", err)
	}
	defer db.Close()

	s := mcpServer.NewServer(db, cfg)

	if err := s.ServeStdio(); err != nil {
		log.Fatalf("Serve stdio: %v", err)
	}
}
