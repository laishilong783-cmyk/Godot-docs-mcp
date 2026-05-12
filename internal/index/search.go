package index

import (
	"database/sql"
	"fmt"
	"strings"
)

// SearchResult represents a single search result.
type SearchResult struct {
	Title   string  `json:"title"`
	Path    string  `json:"path"`
	Section string  `json:"section"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet"`
}

// Search performs a full-text search on documents.
func (db *DB) Search(version, query string, limit int) ([]SearchResult, error) {
	// Escape FTS5 special chars except * and " and space.
	escaped := escapeFTS5(query)

	rows, err := db.conn.Query(`
		SELECT d.title, d.path, d.section, rank, snippet(documents_fts, 2, '<b>', '</b>', '...', 32) as snippet
		FROM documents_fts
		JOIN documents d ON d.id = documents_fts.rowid
		WHERE documents_fts MATCH ? AND d.version = ?
		ORDER BY rank
		LIMIT ?
	`, escaped, version, limit)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var rank float64
		if err := rows.Scan(&r.Title, &r.Path, &r.Section, &rank, &r.Snippet); err != nil {
			continue
		}
		r.Score = -rank // bm25 rank is negative, lower is better
		if r.Score < 0 {
			r.Score = -r.Score
		}
		// Clean snippet tags for plain text.
		r.Snippet = strings.ReplaceAll(r.Snippet, "<b>", "")
		r.Snippet = strings.ReplaceAll(r.Snippet, "</b>", "")
		results = append(results, r)
	}

	return results, nil
}

// escapeFTS5 escapes special characters for FTS5 MATCH.
func escapeFTS5(s string) string {
	// FTS5 special chars: " * ( ) { } [ ] ^ ~ \ : - = < > ` + / . , ; ! @ # $ % & |
	// We keep * for wildcards, quotes for phrase search.
	// Strategy: split by space, quote each token.
	tokens := strings.Fields(s)
	for i, t := range tokens {
		tokens[i] = `"` + strings.ReplaceAll(t, `"`, `""`) + `"`
	}
	return strings.Join(tokens, " ")
}

// GetPage retrieves a document page by path.
func (db *DB) GetPage(version, path string) (title, section, content string, err error) {
	err = db.conn.QueryRow(
		"SELECT title, section, content FROM documents WHERE version = ? AND path = ?",
		version, path,
	).Scan(&title, &section, &content)
	if err == sql.ErrNoRows {
		return "", "", "", fmt.Errorf("page not found: %s", path)
	}
	return title, section, content, err
}

// GetClass retrieves class info by name (case-insensitive).
func (db *DB) GetClass(version, className string) (map[string]interface{}, error) {
	row := db.conn.QueryRow(`
		SELECT path, description FROM symbols
		WHERE version = ? AND kind = 'class' AND class_name LIKE ?
		LIMIT 1
	`, version, className)

	var path, description string
	if err := row.Scan(&path, &description); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("CLASS_NOT_FOUND")
		}
		return nil, err
	}

	// Find inheritance from content.
	_, _, content, err := db.GetPage(version, path)
	if err != nil {
		content = ""
	}

	inherits := extractInherits(content)
	summary := description
	if summary == "" {
		summary = extractSummary(content)
	}

	return map[string]interface{}{
		"class_name": className,
		"version":    version,
		"path":       path,
		"inherits":   inherits,
		"summary":    summary,
	}, nil
}

// GetClassMembers retrieves all members of a class.
func (db *DB) GetClassMembers(version, className string) ([]map[string]interface{}, error) {
	rows, err := db.conn.Query(`
		SELECT kind, member_name, signature, return_type, description
		FROM symbols
		WHERE version = ? AND class_name LIKE ? AND kind != 'class'
		ORDER BY kind, member_name
	`, version, className)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var kind, memberName, signature, returnType, description string
		if err := rows.Scan(&kind, &memberName, &signature, &returnType, &description); err != nil {
			continue
		}
		members = append(members, map[string]interface{}{
			"kind":        kind,
			"name":        memberName,
			"signature":   signature,
			"return_type": returnType,
			"description": description,
		})
	}

	return members, nil
}

// GetMethod retrieves a specific method by class and method name.
func (db *DB) GetMethod(version, className, methodName string) (map[string]interface{}, error) {
	row := db.conn.QueryRow(`
		SELECT signature, return_type, description, path
		FROM symbols
		WHERE version = ? AND class_name LIKE ? AND member_name LIKE ? AND kind = 'method'
		LIMIT 1
	`, version, className, methodName)

	var signature, returnType, description, path string
	if err := row.Scan(&signature, &returnType, &description, &path); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("METHOD_NOT_FOUND")
		}
		return nil, err
	}

	return map[string]interface{}{
		"class_name":  className,
		"method_name": methodName,
		"version":     version,
		"signature":   signature,
		"return_type": returnType,
		"description": description,
		"path":        path,
	}, nil
}

// FindCandidates returns similar class or member names.
func (db *DB) FindCandidates(version, kind, className, memberName string, limit int) ([]string, error) {
	var rows *sql.Rows
	var err error

	if memberName == "" {
		// Find similar classes.
		rows, err = db.conn.Query(`
			SELECT class_name FROM symbols
			WHERE version = ? AND kind = 'class' AND class_name LIKE ?
			LIMIT ?
		`, version, "%"+className+"%", limit)
	} else {
		// Find similar members in the same class or across classes.
		rows, err = db.conn.Query(`
			SELECT member_name FROM symbols
			WHERE version = ? AND class_name LIKE ? AND member_name LIKE ? AND kind = ?
			LIMIT ?
		`, version, className, "%"+memberName+"%", kind, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []string
	seen := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		if !seen[name] {
			seen[name] = true
			candidates = append(candidates, name)
		}
	}

	return candidates, nil
}

// extractInherits extracts inheritance info from RST content.
func extractInherits(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "inherits:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "inherits:")+strings.TrimPrefix(line, "Inherits:"))
		}
		if strings.HasPrefix(strings.ToLower(line), "extends ") {
			return strings.TrimSpace(strings.TrimPrefix(strings.ToLower(line), "extends "))
		}
	}
	return ""
}

// extractSummary extracts a short summary from content.
func extractSummary(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && len(line) > 20 {
			return line
		}
	}
	return ""
}
