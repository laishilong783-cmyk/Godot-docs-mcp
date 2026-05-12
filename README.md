# godot-docs-mcp

A local documentation MCP Server for Godot Engine.

## Why godot-docs-mcp?

When using `godot-mcp-pro` to operate Godot projects, AI Agents sometimes invent APIs, use outdated syntax, or confuse Godot 3.x with 4.x. This MCP Server provides a local, offline, factual check layer by indexing the official Godot documentation and exposing search and query tools.

**Recommended workflow:** Query docs first, then operate Godot.

## Prerequisites

- Go 1.21+
- A local clone of the [Godot docs repository](https://github.com/godotengine/godot-docs)

## Download Godot Official Docs

```bash
git clone https://github.com/godotengine/godot-docs.git
cd godot-docs
git checkout 4.4  # or your target version
```

## Cross-Platform Builds

The project includes pre-built binaries for Windows, macOS, and Linux.

| Platform | Architecture | MCP Server | Indexer |
|----------|-------------|------------|---------|
| Windows | amd64 | `godot-docs-mcp.exe` | `godot-docs-index.exe` |
| macOS (Intel) | amd64 | `godot-docs-mcp-darwin-amd64` | `godot-docs-index-darwin-amd64` |
| macOS (Apple Silicon) | arm64 | `godot-docs-mcp-darwin-arm64` | `godot-docs-index-darwin-arm64` |
| Linux | amd64 | `godot-docs-mcp-linux-amd64` | `godot-docs-index-linux-amd64` |
| Linux | arm64 | `godot-docs-mcp-linux-arm64` | `godot-docs-index-linux-arm64` |

### macOS / Linux 添加执行权限

```bash
chmod +x godot-docs-mcp-darwin-arm64
chmod +x godot-docs-index-darwin-arm64
```

### Build from source

```bash
go build ./cmd/godot-docs-index
go build ./cmd/godot-docs-mcp
```

## Build Index

```bash
# Windows PowerShell
$env:GODOT_DOCS_PATH = "C:\Users\you\dev\godot-docs"
.\godot-docs-index --docs-path $env:GODOT_DOCS_PATH --version 4.4

# Linux/macOS
export GODOT_DOCS_PATH=/Users/you/dev/godot-docs
./godot-docs-index --docs-path $GODOT_DOCS_PATH --version 4.4
```

## Configure MCP Client

### Windows

```json
{
  "mcpServers": {
    "godot-docs": {
      "command": "E:/godot/godot-docs-mcp/godot-docs-mcp.exe",
      "args": [],
      "env": {
        "GODOT_DOCS_PATH": "E:/godot/godot-docs",
        "GODOT_DOCS_DB": "E:/godot/godot_docs_4_4.db",
        "GODOT_DOCS_VERSION": "4.4"
      }
    }
  }
}
```

### macOS (Apple Silicon / Intel)

```json
{
  "mcpServers": {
    "godot-docs": {
      "command": "/Users/you/dev/godot-docs-mcp/godot-docs-mcp-darwin-arm64",
      "args": [],
      "env": {
        "GODOT_DOCS_PATH": "/Users/you/dev/godot-docs",
        "GODOT_DOCS_DB": "/Users/you/dev/godot_docs_4_4.db",
        "GODOT_DOCS_VERSION": "4.4"
      }
    }
  }
}
```

> **Note:** Replace `darwin-arm64` with `darwin-amd64` if you are on Intel Mac.

### Linux

```json
{
  "mcpServers": {
    "godot-docs": {
      "command": "/home/you/dev/godot-docs-mcp/godot-docs-mcp-linux-amd64",
      "args": [],
      "env": {
        "GODOT_DOCS_PATH": "/home/you/dev/godot-docs",
        "GODOT_DOCS_DB": "/home/you/dev/godot_docs_4_4.db",
        "GODOT_DOCS_VERSION": "4.4"
      }
    }
  }
}
```

If you also use `godot-mcp-pro`, keep both servers in the config.

## MCP Tools

| Tool | Description |
|------|-------------|
| `godot_docs_search` | Full-text search across documentation |
| `godot_docs_get_page` | Read a document page by relative path |
| `godot_docs_get_class` | Query class info, inheritance, methods, properties, signals |
| `godot_docs_get_method` | Query method signature and description |

## Example Queries

**Search:**
```json
{"query": "CharacterBody2D move_and_slide", "version": "4.4", "limit": 10}
```

**Get class:**
```json
{"class_name": "CharacterBody2D", "version": "4.4", "include_members": true}
```

**Get method:**
```json
{"class_name": "CharacterBody2D", "method_name": "move_and_slide", "version": "4.4"}
```

## Recommended Workflow with godot-mcp-pro

1. Agent receives a Godot task.
2. Agent calls `godot_docs_search` or `godot_docs_get_class` to verify APIs.
3. Agent calls `godot_docs_get_method` to confirm method signatures.
4. Agent calls `godot-mcp-pro` to modify the project.
5. Agent runs the project and queries docs again if errors occur.

## Common Errors

| Error | Fix |
|-------|-----|
| `GODOT_DOCS_PATH not set` | Set the environment variable to your local docs clone. |
| `Index database not found` | Run `godot-docs-index` first. |
| `Class not found` | Check version, spelling, or run index again. |
| `Invalid path` | Only `.rst`, `.md`, `.txt` files within `GODOT_DOCS_PATH` are allowed. |
