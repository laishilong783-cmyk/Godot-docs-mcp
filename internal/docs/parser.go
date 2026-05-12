package docs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	return &ParsedDoc{
		Title:   title,
		Path:    doc.Path,
		Section: doc.Section,
		Content: content,
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
	KindClass         SymbolKind = "class"
	KindMethod        SymbolKind = "method"
	KindProperty      SymbolKind = "property"
	KindSignal        SymbolKind = "signal"
	KindEnum          SymbolKind = "enum"
	KindConstant      SymbolKind = "constant"
	KindAnnotation    SymbolKind = "annotation"
	KindOperator      SymbolKind = "operator"
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

	// Parse members from raw RST using classref markers.
	symbols = append(symbols, parseMembersFromRawRST(parsed.Content, className, doc.Path)...)

	return parsed, symbols, nil
}

// parseMembersFromRawRST parses Godot 4.4+ classref RST format.
func parseMembersFromRawRST(content, className, path string) []Symbol {
	var symbols []Symbol
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentKind SymbolKind
	var currentMember Symbol
	var descLines []string
	var inDesc bool
	var sigLine string
	lineNum := 0

	flushMember := func() {
		if currentMember.MemberName != "" {
			currentMember.Description = strings.TrimSpace(strings.Join(descLines, "\n"))
			if len(currentMember.Description) > 2000 {
				currentMember.Description = currentMember.Description[:2000] + "..."
			}
			symbols = append(symbols, currentMember)
		}
		currentMember = Symbol{}
		descLines = nil
		inDesc = false
		sigLine = ""
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect classref markers (Godot 4.4+ format)
		switch {
		case strings.Contains(trimmed, ".. rst-class:: classref-method"):
			flushMember()
			currentKind = KindMethod
			continue
		case strings.Contains(trimmed, ".. rst-class:: classref-property"):
			flushMember()
			currentKind = KindProperty
			continue
		case strings.Contains(trimmed, ".. rst-class:: classref-signal"):
			flushMember()
			currentKind = KindSignal
			continue
		case strings.Contains(trimmed, ".. rst-class:: classref-enumeration"):
			flushMember()
			currentKind = KindEnum
			continue
		case strings.Contains(trimmed, ".. rst-class:: classref-constant"):
			flushMember()
			currentKind = KindConstant
			continue
		case strings.Contains(trimmed, ".. rst-class:: classref-annotation"):
			flushMember()
			currentKind = KindAnnotation
			continue
		case strings.Contains(trimmed, ".. rst-class:: classref-operator"):
			flushMember()
			currentKind = KindOperator
			continue
		case strings.Contains(trimmed, ".. rst-class:: classref-themeitem"):
			flushMember()
			currentKind = KindThemeProperty
			continue
		}

		if currentKind == "" {
			continue
		}

		// Signature line: first non-empty line after classref marker
		if sigLine == "" && trimmed != "" {
			sigLine = trimmed
			currentMember = parseClassrefSignature(sigLine, currentKind, className, path, lineNum)
			continue
		}

		// Empty line after signature -> description starts next
		if sigLine != "" && trimmed == "" && !inDesc {
			inDesc = true
			continue
		}

		// Collect description lines
		if inDesc {
			// Stop at next classref marker, section header, or property/signal group separator
			if isSectionHeader(trimmed) || strings.HasPrefix(trimmed, ".. rst-class::") {
				flushMember()
				currentKind = ""
				continue
			}
			// Also stop at table row separators (properties tables)
			if strings.HasPrefix(trimmed, "+") && strings.HasSuffix(trimmed, "+") {
				// Table separator line - might indicate end of description
				if len(descLines) > 0 {
					flushMember()
					currentKind = ""
					continue
				}
			}
			descLines = append(descLines, line)
		}
	}
	flushMember()

	return symbols
}

// parseClassrefSignature parses a Godot 4.4 classref signature line.
func parseClassrefSignature(line string, kind SymbolKind, className, path string, lineNum int) Symbol {
	memberName := extractBoldText(line)
	returnType := extractReturnTypeFromClassref(line, kind)

	// Clean signature for display: remove link refs and escape sequences
	cleanSig := cleanSignatureForDisplay(line)

	return Symbol{
		Kind:       kind,
		ClassName:  className,
		MemberName: memberName,
		Signature:  cleanSig,
		ReturnType: returnType,
		Path:       path,
		LineStart:  lineNum,
	}
}

// extractBoldText extracts text from **bold** RST markup.
func extractBoldText(line string) string {
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		name := strings.ReplaceAll(matches[1], `\`, "")
		return strings.TrimSpace(name)
	}
	return ""
}

// extractReturnTypeFromClassref extracts the return type from a Godot classref signature line.
func extractReturnTypeFromClassref(line string, kind SymbolKind) string {
	// Signals and theme items typically have no return type.
	if kind == KindSignal || kind == KindThemeProperty {
		return ""
	}

	// Check for |void| marker (common for setters and some methods).
	if strings.Contains(line, "|void|") {
		return "void"
	}

	// For properties/methods, the type is the first :ref: before the **bold** member name.
	boldIdx := strings.Index(line, "**")
	if boldIdx == -1 {
		return ""
	}

	re := regexp.MustCompile(":ref:`([^`]+)`")
	matches := re.FindAllStringSubmatchIndex(line, -1)
	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		// Only consider refs that appear before the bold member name.
		if m[0] >= boldIdx {
			continue
		}
		ref := line[m[2]:m[3]]
		// Skip anchor/link emoji refs.
		if strings.Contains(ref, "🔗") {
			continue
		}
		// Extract type from "Type<class_Type>" or "Type<enum_Type>".
		parts := strings.SplitN(ref, "<", 2)
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	return ""
}

// cleanSignatureForDisplay cleans a signature line for human readability.
func cleanSignatureForDisplay(line string) string {
	// Remove :ref:`🔗<...>` anchor links
	re := regexp.MustCompile("\\s*:ref:`🔗[^`]*`")
	line = re.ReplaceAllString(line, "")

	// Simplify :ref:`Type<class_Type>` to Type
	re = regexp.MustCompile(":ref:`([^`]+)`")
	line = re.ReplaceAllStringFunc(line, func(s string) string {
		inner := s[6 : len(s)-1] // remove :ref:` and `
		parts := strings.SplitN(inner, "<", 2)
		return parts[0]
	})

	// Remove |void| wrapper
	line = strings.ReplaceAll(line, "|void|", "void")

	// Remove ** around method names
	re = regexp.MustCompile("\\*\\*([^*]+)\\*\\*")
	line = re.ReplaceAllStringFunc(line, func(s string) string {
		return s[2 : len(s)-2]
	})

	// Unescape RST backslashes (Godot RST uses \ before special chars or words)
	line = strings.ReplaceAll(line, `\ (`, "(")
	line = strings.ReplaceAll(line, `\ )`, ")")
	line = strings.ReplaceAll(line, `\:`, ":")
	line = strings.ReplaceAll(line, `\,`, ",")
	line = strings.ReplaceAll(line, `\=`, "=")
	line = strings.ReplaceAll(line, `\[`, "[")
	line = strings.ReplaceAll(line, `\]`, "]")
	line = strings.ReplaceAll(line, `\.`, ".")
	// Remove remaining backslash-space patterns (e.g. \ param_name)
	line = strings.ReplaceAll(line, `\ `, " ")

	return strings.TrimSpace(line)
}

// isSectionHeader detects RST section header underlines.
func isSectionHeader(line string) bool {
	if len(line) < 3 {
		return false
	}
	first := rune(line[0])
	if first != '=' && first != '-' && first != '~' && first != '^' && first != '*' {
		return false
	}
	for _, r := range line {
		if r != first {
			return false
		}
	}
	return true
}

// --- Legacy parsers kept for compatibility ---

func parseMembers(content, className, path string) []Symbol {
	// Fallback: old parser for non-classref RST content.
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
			if len(currentMember.Description) > 2000 {
				currentMember.Description = currentMember.Description[:2000] + "..."
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

		if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") {
			flushMember()
			currentMember = parseMethodSignature(trimmed, className, path, lineNum)
			inMemberDesc = true
			continue
		}

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
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return Symbol{}
	}

	returnType := ""
	nameStart := 0
	if isTypeLike(parts[0]) || parts[0] == "static" || parts[0] == "const" || parts[0] == "virtual" || parts[0] == "abstract" || parts[0] == "override" {
		if parts[0] != "static" && parts[0] != "const" && parts[0] != "virtual" && parts[0] != "abstract" && parts[0] != "override" {
			returnType = parts[0]
		}
		nameStart = 1
	}

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

func isTypeLike(s string) bool {
	if s == "" {
		return false
	}
	first := rune(s[0])
	return unicode.IsUpper(first) || s == "void" || s == "bool" || s == "int" || s == "float" || s == "String" || s == "Variant" || s == "null"
}
