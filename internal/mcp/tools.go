package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/godot-docs-mcp/godot-docs-mcp/internal/config"
	"github.com/godot-docs-mcp/godot-docs-mcp/internal/index"
)

// Server wraps the MCP server with our index and config.
type Server struct {
	s   *server.MCPServer
	db  *index.DB
	cfg *config.Config
}

// NewServer creates a new MCP server instance.
func NewServer(db *index.DB, cfg *config.Config) *Server {
	s := server.NewMCPServer(
		"godot-docs-mcp",
		"1.0.0",
		server.WithResourceCapabilities(false, false),
	)

	ms := &Server{s: s, db: db, cfg: cfg}
	ms.registerTools()
	return ms
}

// ServeStdio starts the MCP server over stdio.
func (ms *Server) ServeStdio() error {
	return server.ServeStdio(ms.s)
}

func (ms *Server) registerTools() {
	// Tool: godot_docs_search
	searchTool := mcp.NewTool("godot_docs_search",
		mcp.WithDescription("Full-text search local Godot documentation. Use when you are unsure of the exact class name, method name, or tutorial location."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query text")),
		mcp.WithString("version", mcp.Description("Godot docs version, e.g. 4.4")),
		mcp.WithNumber("limit", mcp.Description("Max number of results (default 10, max 50)")),
	)
	ms.s.AddTool(searchTool, ms.handleSearch)

	// Tool: godot_docs_get_page
	pageTool := mcp.NewTool("godot_docs_get_page",
		mcp.WithDescription("Read a document page by relative path."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path to the document under GODOT_DOCS_PATH")),
		mcp.WithString("version", mcp.Description("Godot docs version, e.g. 4.4")),
		mcp.WithNumber("max_chars", mcp.Description("Maximum characters to return")),
	)
	ms.s.AddTool(pageTool, ms.handleGetPage)

	// Tool: godot_docs_get_class
	classTool := mcp.NewTool("godot_docs_get_class",
		mcp.WithDescription("Query Godot class documentation. Use member_kinds, member_query and member_limit to avoid returning huge class member lists."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("class_name", mcp.Required(), mcp.Description("Class name, e.g. CharacterBody2D")),
		mcp.WithString("version", mcp.Description("Godot docs version, e.g. 4.4")),
		mcp.WithBoolean("include_members", mcp.Description("Whether to include class members. Prefer using member filters to reduce token usage.")),
		mcp.WithArray("member_kinds", mcp.Description("Optional member kinds to include, e.g. [\"method\", \"property\", \"signal\"]"), mcp.WithStringItems()),
		mcp.WithString("member_query", mcp.Description("Optional member name/signature filter, e.g. slide or velocity")),
		mcp.WithNumber("member_limit", mcp.Description("Max included members (default 50, max 200)")),
	)
	ms.s.AddTool(classTool, ms.handleGetClass)

	// Tool: godot_docs_get_method
	methodTool := mcp.NewTool("godot_docs_get_method",
		mcp.WithDescription("Query a method signature and description of a Godot class. Use to avoid inventing method parameters or return types."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("class_name", mcp.Required(), mcp.Description("Class name, e.g. CharacterBody2D")),
		mcp.WithString("method_name", mcp.Required(), mcp.Description("Method name, e.g. move_and_slide")),
		mcp.WithString("version", mcp.Description("Godot docs version, e.g. 4.4")),
	)
	ms.s.AddTool(methodTool, ms.handleGetMethod)

	// Tool: godot_docs_get_property
	propertyTool := mcp.NewTool("godot_docs_get_property",
		mcp.WithDescription("Query a Godot class property type and description without returning the whole class."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("class_name", mcp.Required(), mcp.Description("Class name, e.g. CharacterBody2D")),
		mcp.WithString("property_name", mcp.Required(), mcp.Description("Property name, e.g. velocity")),
		mcp.WithString("version", mcp.Description("Godot docs version, e.g. 4.4")),
	)
	ms.s.AddTool(propertyTool, ms.handleGetProperty)

	// Tool: godot_docs_get_signal
	signalTool := mcp.NewTool("godot_docs_get_signal",
		mcp.WithDescription("Query a Godot class signal signature and description without returning the whole class."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("class_name", mcp.Required(), mcp.Description("Class name, e.g. Area2D")),
		mcp.WithString("signal_name", mcp.Required(), mcp.Description("Signal name, e.g. body_entered")),
		mcp.WithString("version", mcp.Description("Godot docs version, e.g. 4.4")),
	)
	ms.s.AddTool(signalTool, ms.handleGetSignal)

	// Tool: godot_docs_suggest_apis
	suggestTool := mcp.NewTool("godot_docs_suggest_apis",
		mcp.WithDescription("Return a compact list of likely relevant Godot API symbols for a task. Use before calling godot-mcp-pro."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("task", mcp.Required(), mcp.Description("Task or API keywords, e.g. CharacterBody2D velocity move_and_slide")),
		mcp.WithString("version", mcp.Description("Godot docs version, e.g. 4.4")),
		mcp.WithArray("kinds", mcp.Description("Optional symbol kinds, e.g. [\"method\", \"property\", \"signal\"]"), mcp.WithStringItems()),
		mcp.WithNumber("limit", mcp.Description("Max API symbols to return (default 12, max 50)")),
	)
	ms.s.AddTool(suggestTool, ms.handleSuggestAPIs)
}

func (ms *Server) handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	query, _ := args["query"].(string)
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	version := ms.cfg.Version
	if v, ok := args["version"].(string); ok && v != "" {
		version = v
	}

	limit := ms.cfg.MaxResults
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	if limit > 50 {
		limit = 50
	}

	results, err := ms.db.Search(version, query, limit)
	if err != nil {
		return errorResult("SEARCH_ERROR", err.Error(), ""), nil
	}

	resp := map[string]interface{}{
		"query":   query,
		"version": version,
		"results": results,
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (ms *Server) handleGetPage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	path, _ := args["path"].(string)
	if path == "" {
		return errorResult("INVALID_PATH", "path is required", ""), nil
	}

	// Security: validate path.
	if err := validatePath(ms.cfg.DocsPath, path); err != nil {
		return errorResult("INVALID_PATH", err.Error(), ""), nil
	}

	version := ms.cfg.Version
	if v, ok := args["version"].(string); ok && v != "" {
		version = v
	}

	maxChars := ms.cfg.MaxContentChars
	if m, ok := args["max_chars"].(float64); ok && m > 0 {
		maxChars = int(m)
	}

	title, section, content, err := ms.db.GetPage(version, filepath.ToSlash(path))
	if err != nil {
		return errorResult("PAGE_NOT_FOUND", err.Error(), "Check the path and run godot-docs-index if needed."), nil
	}

	truncated := false
	if len(content) > maxChars {
		content = content[:maxChars]
		truncated = true
	}

	resp := map[string]interface{}{
		"title":         title,
		"path":          path,
		"version":       version,
		"section":       section,
		"content":       content,
		"truncated":     truncated,
		"content_chars": len(content),
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (ms *Server) handleGetClass(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	className, _ := args["class_name"].(string)
	if className == "" {
		return errorResult("INVALID_ARGUMENT", "class_name is required", ""), nil
	}

	version := ms.cfg.Version
	if v, ok := args["version"].(string); ok && v != "" {
		version = v
	}

	includeMembers := false
	if inc, ok := args["include_members"].(bool); ok {
		includeMembers = inc
	}
	memberKinds := stringSliceArg(args, "member_kinds")
	memberQuery, _ := args["member_query"].(string)
	memberLimit := intArg(args, "member_limit", 50, 200)
	_, hasMemberLimit := args["member_limit"]
	if len(memberKinds) > 0 || memberQuery != "" || hasMemberLimit {
		includeMembers = true
	}

	classInfo, err := ms.db.GetClass(version, className)
	if err != nil {
		if err.Error() == "CLASS_NOT_FOUND" {
			candidates, _ := ms.db.FindCandidates(version, "class", className, "", 5)
			return errorResult("CLASS_NOT_FOUND",
				fmt.Sprintf("Class '%s' was not found in Godot docs version %s.", className, version),
				fmt.Sprintf("Candidates: %v", candidates)), nil
		}
		return errorResult("SEARCH_ERROR", err.Error(), ""), nil
	}

	if includeMembers {
		members, err := ms.db.GetClassMembers(version, className, index.MemberFilter{
			Kinds: memberKinds,
			Query: memberQuery,
			Limit: memberLimit,
		})
		if err == nil {
			// Group by kind.
			methods := []map[string]interface{}{}
			properties := []map[string]interface{}{}
			signals := []map[string]interface{}{}
			for _, m := range members {
				switch m["kind"] {
				case "method":
					methods = append(methods, map[string]interface{}{"name": m["name"], "signature": m["signature"], "return_type": m["return_type"]})
				case "property":
					properties = append(properties, map[string]interface{}{"name": m["name"], "type": m["return_type"]})
				case "signal":
					signals = append(signals, map[string]interface{}{"name": m["name"], "signature": m["signature"]})
				}
			}
			classInfo["methods"] = methods
			classInfo["properties"] = properties
			classInfo["signals"] = signals
			classInfo["member_filter"] = map[string]interface{}{
				"kinds": memberKinds,
				"query": memberQuery,
				"limit": memberLimit,
			}
		}
	}

	data, _ := json.MarshalIndent(classInfo, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (ms *Server) handleGetMethod(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	className, _ := args["class_name"].(string)
	methodName, _ := args["method_name"].(string)
	if className == "" || methodName == "" {
		return errorResult("INVALID_ARGUMENT", "class_name and method_name are required", ""), nil
	}

	version := ms.cfg.Version
	if v, ok := args["version"].(string); ok && v != "" {
		version = v
	}

	methodInfo, err := ms.db.GetMethod(version, className, methodName)
	if err != nil {
		if err.Error() == "METHOD_NOT_FOUND" {
			candidates, _ := ms.db.FindCandidates(version, "method", className, methodName, 5)
			return errorResult("METHOD_NOT_FOUND",
				fmt.Sprintf("Method '%s' was not found on class '%s' in Godot docs version %s.", methodName, className, version),
				fmt.Sprintf("Candidates: %v", candidates)), nil
		}
		return errorResult("SEARCH_ERROR", err.Error(), ""), nil
	}

	data, _ := json.MarshalIndent(methodInfo, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (ms *Server) handleGetProperty(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	className, _ := args["class_name"].(string)
	propertyName, _ := args["property_name"].(string)
	if className == "" || propertyName == "" {
		return errorResult("INVALID_ARGUMENT", "class_name and property_name are required", ""), nil
	}

	version := ms.cfg.Version
	if v, ok := args["version"].(string); ok && v != "" {
		version = v
	}

	info, err := ms.db.GetMember(version, className, "property", propertyName)
	if err != nil {
		if err.Error() == "MEMBER_NOT_FOUND" {
			candidates, _ := ms.db.FindCandidates(version, "property", className, propertyName, 5)
			return errorResult("PROPERTY_NOT_FOUND",
				fmt.Sprintf("Property '%s' was not found on class '%s' in Godot docs version %s.", propertyName, className, version),
				fmt.Sprintf("Candidates: %v", candidates)), nil
		}
		return errorResult("SEARCH_ERROR", err.Error(), ""), nil
	}
	info["property_name"] = propertyName

	data, _ := json.MarshalIndent(info, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (ms *Server) handleGetSignal(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	className, _ := args["class_name"].(string)
	signalName, _ := args["signal_name"].(string)
	if className == "" || signalName == "" {
		return errorResult("INVALID_ARGUMENT", "class_name and signal_name are required", ""), nil
	}

	version := ms.cfg.Version
	if v, ok := args["version"].(string); ok && v != "" {
		version = v
	}

	info, err := ms.db.GetMember(version, className, "signal", signalName)
	if err != nil {
		if err.Error() == "MEMBER_NOT_FOUND" {
			candidates, _ := ms.db.FindCandidates(version, "signal", className, signalName, 5)
			return errorResult("SIGNAL_NOT_FOUND",
				fmt.Sprintf("Signal '%s' was not found on class '%s' in Godot docs version %s.", signalName, className, version),
				fmt.Sprintf("Candidates: %v", candidates)), nil
		}
		return errorResult("SEARCH_ERROR", err.Error(), ""), nil
	}
	info["signal_name"] = signalName

	data, _ := json.MarshalIndent(info, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (ms *Server) handleSuggestAPIs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	task, _ := args["task"].(string)
	if strings.TrimSpace(task) == "" {
		return errorResult("INVALID_ARGUMENT", "task is required", ""), nil
	}

	version := ms.cfg.Version
	if v, ok := args["version"].(string); ok && v != "" {
		version = v
	}
	limit := intArg(args, "limit", 12, 50)
	kinds := stringSliceArg(args, "kinds")

	symbols, err := ms.db.SearchSymbols(version, task, kinds, limit)
	if err != nil {
		return errorResult("SEARCH_ERROR", err.Error(), ""), nil
	}

	docLimit := limit / 3
	if docLimit < 3 {
		docLimit = 3
	}
	if docLimit > 8 {
		docLimit = 8
	}
	pages, _ := ms.db.Search(version, task, docLimit)

	resp := map[string]interface{}{
		"task":          task,
		"version":       version,
		"api_symbols":   symbols,
		"related_pages": pages,
		"usage":         "Query exact methods/properties/signals before modifying Godot through godot-mcp-pro.",
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// validatePath ensures the relative path is safe and within GODOT_DOCS_PATH.
func validatePath(docsPath, relPath string) error {
	relPath = filepath.Clean(relPath)
	if strings.Contains(relPath, "..") {
		return fmt.Errorf("path cannot contain path traversal")
	}
	ext := strings.ToLower(filepath.Ext(relPath))
	if ext != ".rst" && ext != ".md" && ext != ".txt" {
		return fmt.Errorf("only .rst, .md, and .txt files are allowed")
	}
	fullPath := filepath.Join(docsPath, relPath)
	absDocs, _ := filepath.Abs(docsPath)
	absFull, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFull, absDocs) {
		return fmt.Errorf("path must be within GODOT_DOCS_PATH")
	}
	return nil
}

func intArg(args map[string]interface{}, name string, fallback, max int) int {
	n := fallback
	if v, ok := args[name].(float64); ok && v > 0 {
		n = int(v)
	}
	if n > max {
		return max
	}
	return n
}

func stringSliceArg(args map[string]interface{}, name string) []string {
	raw, ok := args[name].([]interface{})
	if !ok {
		return nil
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok && s != "" {
			values = append(values, s)
		}
	}
	return values
}

func errorResult(code, message, hint string) *mcp.CallToolResult {
	resp := map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
			"hint":    hint,
		},
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return mcp.NewToolResultText(string(data))
}
