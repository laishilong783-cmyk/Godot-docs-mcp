package docs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// ParsedDoc holds parsed document info.
type ParsedDoc struct {
	Title   string
	Path    string
	Section string
	Content string
}

// ParseFile parses a single document file.
func ParseFile(docsPath string, doc DocFile) (*ParsedDoc, error) {
	fullPath := filepath.Join(docsPath, doc.Path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	title := extractTitle(content)
	if title == "" {
		title = filepath.Base(doc.Path)
	}

	// Strip RST directives and markup for plain text content.
	plain := stripRST(content)

	return &ParsedDoc{
		Title:   title,
		Path:    doc.Path,
		Section: doc.Section,
		Content: plain,
	}, nil
}

// extractTitle extracts the document title from RST content.
func extractTitle(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var prevLine string
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "..") {
			continue
		}
		if isRSTUnderline(trimmed) && prevLine != "" {
			return strings.TrimSpace(prevLine)
		}
		prevLine = trimmed
	}
	// Fallback: first non-empty line.
	scanner = bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "..") {
			return line
		}
	}
	return ""
}

// isRSTUnderline checks if a line is a valid RST title underline.
func isRSTUnderline(line string) bool {
	if len(line) < 3 {
		return false
	}
	first := rune(line[0])
	if first != '=' && first != '-' && first != '~' && first != '^' && first != '*' && first != '#' && first != '"' {
		return false
	}
	for _, r := range line {
		if r != first {
			return false
		}
	}
	return true
}

// stripRST removes RST markup and directives to produce plain text.
func stripRST(content string) string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	inDirective := false
	inCodeBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip comments.
		if strings.HasPrefix(trimmed, "..") && !strings.HasPrefix(trimmed, "...") {
			// Could be a directive.
			if strings.Contains(trimmed, "::") {
				inDirective = true
			}
			continue
		}

		// Code blocks.
		if strings.HasSuffix(trimmed, "::") {
			inCodeBlock = true
			lines = append(lines, strings.TrimSuffix(line, "::"))
			continue
		}

		if inCodeBlock {
			if trimmed == "" || (len(line) > 0 && line[0] == ' ' || line[0] == '\t') {
				continue // skip code block content
			}
			inCodeBlock = false
		}

		if inDirective {
			if trimmed == "" || (len(line) > 0 && line[0] == ' ' || line[0] == '\t') {
				continue
			}
			inDirective = false
		}

		// Skip RST references like :ref:`something`.
		cleaned := cleanInlineRST(line)
		lines = append(lines, cleaned)
	}

	result := strings.Join(lines, "\n")
	// Collapse multiple blank lines.
	result = collapseBlankLines(result)
	return strings.TrimSpace(result)
}

// cleanInlineRST removes inline RST markup.
func cleanInlineRST(line string) string {
	// Remove ``code`` markers -> code
	line = strings.ReplaceAll(line, "``", "`")
	// Remove *emphasis* and **strong** markers.
	line = replaceBalanced(line, '*', "")
	// Remove :role:`text` -> text
	line = removeRoleMarkers(line)
	// Remove `link`_ markers.
	line = replaceBalanced(line, '`', "")
	return line
}

// removeRoleMarkers removes :role:`text` inline RST markers safely.
func removeRoleMarkers(line string) string {
	var result strings.Builder
	i := 0
	for i < len(line) {
		// Find colon that starts a role
		colonIdx := strings.Index(line[i:], ":`")
		if colonIdx == -1 {
			result.WriteString(line[i:])
			break
		}
		colonIdx += i
		// Look backwards for the role start colon
		roleStart := strings.LastIndex(line[i:colonIdx], ":")
		if roleStart != -1 {
			roleStart += i
		} else {
			// No role start found, just skip this colon
			result.WriteString(line[i : colonIdx+1])
			i = colonIdx + 1
			continue
		}
		// Find closing backtick after colonIdx+2
		backtick := strings.Index(line[colonIdx+2:], "`")
		if backtick == -1 {
			result.WriteString(line[i:])
			break
		}
		contentStart := colonIdx + 2
		contentEnd := colonIdx + 2 + backtick
		result.WriteString(line[i:roleStart])
		result.WriteString(line[contentStart:contentEnd])
		i = contentEnd + 1
	}
	return result.String()
}

// replaceBalanced removes balanced characters and keeps the content between them.
func replaceBalanced(s string, char rune, replacement string) string {
	var result strings.Builder
	depth := 0
	for _, r := range s {
		if r == char {
			if depth == 0 {
				depth = 1
				continue
			} else {
				depth = 0
				continue
			}
		}
		// Always write content, only skip the wrapper chars themselves
		result.WriteRune(r)
	}
	return result.String()
}

// collapseBlankLines collapses 3+ blank lines into 2.
func collapseBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	blankCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount <= 2 {
				out = append(out, line)
			}
		} else {
			blankCount = 0
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// ClassFromPath derives a class name from an RST file path.
func ClassFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimPrefix(strings.TrimSuffix(base, ext), "class_")
	return snakeToPascal(name)
}

// snakeToPascal converts snake_case to PascalCase.
func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	var result strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		// Capitalize first letter if it's a letter
		// Capitalize first letter if it's a letter
		for i := 0; i < len(runes); i++ {
			if unicode.IsLetter(runes[i]) {
				runes[i] = unicode.ToUpper(runes[i])
				break
			}
		}
		// For remaining characters: if previous was a digit and current is a letter, uppercase it
		for i := 1; i < len(runes); i++ {
			if i > 0 && unicode.IsDigit(runes[i-1]) && unicode.IsLetter(runes[i]) {
				runes[i] = unicode.ToUpper(runes[i])
			}
		}
		result.WriteString(string(runes))
	}
	return result.String()
}

// SymbolKind represents the type of a Godot API symbol.
type SymbolKind string

const (
	KindClass        SymbolKind = "class"
	KindMethod       SymbolKind = "method"
	KindProperty     SymbolKind = "property"
	KindSignal       SymbolKind = "signal"
	KindEnum         SymbolKind = "enum"
	KindConstant     SymbolKind = "constant"
	KindAnnotation   SymbolKind = "annotation"
	KindOperator     SymbolKind = "operator"
	KindThemeProperty SymbolKind = "theme_property"
)

// Symbol represents a parsed API symbol.
type Symbol struct {
	Kind        SymbolKind
	ClassName   string
	MemberName  string
	Signature   string
	ReturnType  string
	Description string
	Path        string
	LineStart   int
	LineEnd     int
}

// ParseClassDoc parses a Godot class RST document and extracts symbols.
func ParseClassDoc(docsPath string, doc DocFile) (*ParsedDoc, []Symbol, error) {
	parsed, err := ParseFile(docsPath, doc)
	if err != nil {
		return nil, nil, err
	}

	className := ClassFromPath(doc.Path)
	if parsed.Title != "" && !strings.EqualFold(parsed.Title, className) {
		className = parsed.Title
	}

	var symbols []Symbol

	// Add the class itself.
	symbols = append(symbols, Symbol{
		Kind:      KindClass,
		ClassName: className,
		Path:      doc.Path,
	})

	// Parse members from the original content.
	symbols = append(symbols, parseMembers(parsed.Content, className, doc.Path)...)

	return parsed, symbols, nil
}

func parseMembers(content, className, path string) []Symbol {
	var symbols []Symbol
	scanner := bufio.NewScanner(strings.NewReader(content))
	currentSection := ""
	var currentDescLines []string
	var currentMember Symbol
	inMemberDesc := false
	lineNum := 0

	flushMember := func() {
		if currentMember.MemberName != "" {
			currentMember.Description = strings.TrimSpace(strings.Join(currentDescLines, "\n"))
			if currentMember.Description != "" {
				// Truncate description.
				if len(currentMember.Description) > 2000 {
					currentMember.Description = currentMember.Description[:2000] + "..."
				}
			}
			symbols = append(symbols, currentMember)
		}
		currentMember = Symbol{}
		currentDescLines = nil
		inMemberDesc = false
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect section headers.
		lower := strings.ToLower(trimmed)
		switch lower {
		case "methods", "method descriptions":
			currentSection = "methods"
			flushMember()
			continue
		case "properties", "property descriptions":
			currentSection = "properties"
			flushMember()
			continue
		case "signals", "signal descriptions":
			currentSection = "signals"
			flushMember()
			continue
		case "enumerations", "enum descriptions":
			currentSection = "enums"
			flushMember()
			continue
		case "constants", "constant descriptions":
			currentSection = "constants"
			flushMember()
			continue
		case "annotations":
			currentSection = "annotations"
			flushMember()
			continue
		case "theme properties":
			currentSection = "theme_properties"
			flushMember()
			continue
		}

		if currentSection == "" {
			continue
		}

		// Detect member signature lines.
		if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") {
			// Looks like a method signature.
			flushMember()
			currentMember = parseMethodSignature(trimmed, className, path, lineNum)
			inMemberDesc = true
			continue
		}

		// Simple property/signal line: name type
		if currentSection == "properties" || currentSection == "signals" || currentSection == "theme_properties" {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 && !strings.Contains(trimmed, "(") {
				flushMember()
				memberName := parts[0]
				memberType := strings.Join(parts[1:], " ")
				kind := KindProperty
				if currentSection == "signals" {
					kind = KindSignal
				} else if currentSection == "theme_properties" {
					kind = KindThemeProperty
				}
				currentMember = Symbol{
					Kind:       kind,
					ClassName:  className,
					MemberName: memberName,
					Signature:  fmt.Sprintf("%s %s", memberName, memberType),
					ReturnType: memberType,
					Path:       path,
					LineStart:  lineNum,
				}
				inMemberDesc = true
				continue
			}
		}

		if inMemberDesc {
			currentDescLines = append(currentDescLines, line)
		}
	}
	flushMember()

	return symbols
}

func parseMethodSignature(line, className, path string, lineNum int) Symbol {
	trimmed := strings.TrimSpace(line)
	// Try to extract return type and method name.
	// Format examples:
	// bool move_and_slide()
	// void add_child(node: Node, force_readable_name: bool = false)
	// PackedByteArray save_png_to_buffer() const
	// static Vector2 get_gravity()

	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return Symbol{}
	}

	// Check if first word looks like a return type.
	returnType := ""
	nameStart := 0
	if isTypeLike(parts[0]) || parts[0] == "static" || parts[0] == "const" || parts[0] == "virtual" || parts[0] == "abstract" || parts[0] == "override" {
		if parts[0] != "static" && parts[0] != "const" && parts[0] != "virtual" && parts[0] != "abstract" && parts[0] != "override" {
			returnType = parts[0]
		}
		nameStart = 1
	}

	// Find method name before '('.
	rest := strings.Join(parts[nameStart:], " ")
	parenIdx := strings.Index(rest, "(")
	memberName := rest
	if parenIdx > 0 {
		memberName = strings.TrimSpace(rest[:parenIdx])
	}

	return Symbol{
		Kind:       KindMethod,
		ClassName:  className,
		MemberName: memberName,
		Signature:  trimmed,
		ReturnType: returnType,
		Path:       path,
		LineStart:  lineNum,
	}
}

// isTypeLike checks if a word looks like a type name.
func isTypeLike(s string) bool {
	if s == "" {
		return false
	}
	first := rune(s[0])
	return unicode.IsUpper(first) || s == "void" || s == "bool" || s == "int" || s == "float" || s == "String" || s == "Variant" || s == "null"
}
