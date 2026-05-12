package docs

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// DocFile represents a scanned document file.
type DocFile struct {
	Path    string // relative to docs root
	Section string // e.g., classes, tutorials
	Ext     string
}

// Scan scans the docs directory for supported document files.
func Scan(docsPath string) ([]DocFile, error) {
	var files []DocFile

	err := filepath.WalkDir(docsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".rst" && ext != ".md" && ext != ".txt" {
			return nil
		}

		rel, err := filepath.Rel(docsPath, path)
		if err != nil {
			return nil
		}

		section := ""
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) > 0 {
			section = parts[0]
		}

		files = append(files, DocFile{
			Path:    filepath.ToSlash(rel),
			Section: section,
			Ext:     ext,
		})
		return nil
	})

	return files, err
}

// IsAllowedExt checks if the extension is in the whitelist.
func IsAllowedExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".rst", ".md", ".txt":
		return true
	}
	return false
}
