package index

import (
	"database/sql"
	"fmt"
	"regexp"
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

// MemberFilter controls class member retrieval.
type MemberFilter struct {
	Kinds []string
	Query string
	Limit int
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

// GetClassMembers retrieves members of a class, optionally filtered by kind, name and limit.
func (db *DB) GetClassMembers(version, className string, filter MemberFilter) ([]map[string]interface{}, error) {
	where := []string{"version = ?", "class_name = ? COLLATE NOCASE", "kind != 'class'"}
	args := []interface{}{version, className}

	if len(filter.Kinds) > 0 {
		placeholders := make([]string, 0, len(filter.Kinds))
		for _, kind := range filter.Kinds {
			if !isSymbolKind(kind) || kind == "class" {
				continue
			}
			placeholders = append(placeholders, "?")
			args = append(args, kind)
		}
		if len(placeholders) > 0 {
			where = append(where, "kind IN ("+strings.Join(placeholders, ",")+")")
		}
	}
	if filter.Query != "" {
		where = append(where, "(member_name LIKE ? OR signature LIKE ?)")
		like := "%" + filter.Query + "%"
		args = append(args, like, like)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	args = append(args, limit)

	rows, err := db.conn.Query(`
		SELECT kind, member_name, signature, return_type, description
		FROM symbols
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY kind, member_name
		LIMIT ?
	`, args...)
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

// GetMember retrieves a specific symbol by class, kind and member name.
func (db *DB) GetMember(version, className, kind, memberName string) (map[string]interface{}, error) {
	if !isSymbolKind(kind) || kind == "class" {
		return nil, fmt.Errorf("INVALID_SYMBOL_KIND")
	}

	row := db.conn.QueryRow(`
		SELECT signature, return_type, description, path
		FROM symbols
		WHERE version = ? AND class_name = ? COLLATE NOCASE AND member_name = ? COLLATE NOCASE AND kind = ?
		LIMIT 1
	`, version, className, memberName, kind)

	var signature, returnType, description, path string
	if err := row.Scan(&signature, &returnType, &description, &path); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("MEMBER_NOT_FOUND")
		}
		return nil, err
	}

	return map[string]interface{}{
		"class_name":  className,
		"member_name": memberName,
		"kind":        kind,
		"version":     version,
		"signature":   signature,
		"return_type": returnType,
		"description": description,
		"path":        path,
	}, nil
}

// GetMethod retrieves a specific method by class and method name.
func (db *DB) GetMethod(version, className, methodName string) (map[string]interface{}, error) {
	info, err := db.GetMember(version, className, "method", methodName)
	if err != nil {
		if err.Error() == "MEMBER_NOT_FOUND" {
			return nil, fmt.Errorf("METHOD_NOT_FOUND")
		}
		return nil, err
	}
	info["method_name"] = methodName
	return info, nil
}

// SearchSymbols finds API symbols matching a query in names or signatures.
func (db *DB) SearchSymbols(version, query string, kinds []string, limit int) ([]map[string]interface{}, error) {
	where := []string{"version = ?", "kind != 'class'", "(member_name LIKE ? OR class_name LIKE ? OR signature LIKE ?)"}
	like := "%" + query + "%"
	args := []interface{}{version, like, like, like}

	if len(kinds) > 0 {
		placeholders := make([]string, 0, len(kinds))
		for _, kind := range kinds {
			if !isSymbolKind(kind) || kind == "class" {
				continue
			}
			placeholders = append(placeholders, "?")
			args = append(args, kind)
		}
		if len(placeholders) > 0 {
			where = append(where, "kind IN ("+strings.Join(placeholders, ",")+")")
		}
	}

	if limit <= 0 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}
	args = append(args, limit)

	rows, err := db.conn.Query(`
		SELECT kind, class_name, member_name, signature, return_type, path
		FROM symbols
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY
			CASE
				WHEN member_name = ? COLLATE NOCASE THEN 0
				WHEN member_name LIKE ? THEN 1
				WHEN class_name LIKE ? THEN 2
				ELSE 3
			END,
			class_name, member_name
		LIMIT ?
	`, append(args[:len(args)-1], query, query+"%", query+"%", limit)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var kind, className, memberName, signature, returnType, path string
		if err := rows.Scan(&kind, &className, &memberName, &signature, &returnType, &path); err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"kind":        kind,
			"class_name":  className,
			"name":        memberName,
			"signature":   signature,
			"return_type": returnType,
			"path":        path,
		})
	}
	return results, nil
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

func isSymbolKind(kind string) bool {
	switch kind {
	case "class", "method", "property", "signal", "enum", "constant", "annotation", "operator", "theme_property":
		return true
	default:
		return false
	}
}

// extractInherits extracts inheritance info from RST content.
func extractInherits(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		// Godot 4.4 format: **Inherits:** :ref:`ClassName<class_ClassName>` **<** ...
		if strings.HasPrefix(lower, "**inherits:**") || strings.HasPrefix(lower, "inherits:") {
			// Extract class name from first :ref:`ClassName<class_ClassName>`
			refRe := regexp.MustCompile(":ref:`([^`]+)`")
			matches := refRe.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				ref := m[1]
				parts := strings.SplitN(ref, "<", 2)
				if len(parts) > 0 {
					return strings.TrimSpace(parts[0])
				}
			}
			return ""
		}
		if strings.HasPrefix(lower, "extends ") {
			return strings.TrimSpace(strings.TrimPrefix(lower, "extends "))
		}
	}
	return ""
}

// extractSummary extracts a short summary from content.
func extractSummary(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip directives, comments, empty lines, section headers, inherits line
		if line == "" || strings.HasPrefix(line, "..") || strings.HasPrefix(line, ":") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "**inherits:**") || strings.HasPrefix(lower, "inherits:") {
			continue
		}
		if strings.HasPrefix(lower, "description") {
			continue
		}
		// Skip RST underline markers (==== ----)
		if len(line) >= 3 {
			first := line[0]
			if first == '=' || first == '-' || first == '~' || first == '^' || first == '*' {
				allSame := true
				for i := 1; i < len(line); i++ {
					if line[i] != first {
						allSame = false
						break
					}
				}
				if allSame {
					continue
				}
			}
		}
		// Skip very short lines and titles
		if len(line) > 30 {
			return line
		}
	}
	// Fallback: first non-empty, non-comment line
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "..") && !strings.HasPrefix(line, ":") {
			return line
		}
	}
	return ""
}
