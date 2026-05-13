# godot-docs-mcp

English | [中文](#中文)

A local, offline MCP server for querying official Godot documentation before an AI agent modifies a Godot project.

`godot-docs-mcp` is designed to work together with tools such as `godot-mcp-pro`: the agent checks the local Godot docs first, confirms real classes, methods, properties, and signals, then uses the Godot operation MCP to edit the project. This reduces hallucinated APIs, Godot 3.x/4.x mix-ups, and incorrect scene/script generation.

## Contents

- [Features](#features)
- [MCP Tools](#mcp-tools)
- [Prerequisites](#prerequisites)
- [Quick Start: Windows](#quick-start-windows)
- [Quick Start: macOS](#quick-start-macos)
- [Quick Start: Linux](#quick-start-linux)
- [Multiple Godot Versions](#multiple-godot-versions)
- [Example Queries](#example-queries)
- [中文](#中文)

## Features

- Indexes a local clone of the official `godot-docs` repository.
- Stores docs and API symbols in SQLite with full-text search.
- Supports multiple Godot documentation versions in the same database.
- Exposes read-only MCP tools for classes, methods, properties, signals, pages, and compact API suggestions.
- Works offline after the documentation repository has been cloned and indexed.

## MCP Tools

| Tool | Description |
|------|-------------|
| `godot_docs_search` | Full-text search across indexed documentation. |
| `godot_docs_get_page` | Read a document page by relative path. |
| `godot_docs_get_class` | Query class info and optionally filtered members. |
| `godot_docs_get_method` | Query a method signature and description. |
| `godot_docs_get_property` | Query a property type and description. |
| `godot_docs_get_signal` | Query a signal signature and description. |
| `godot_docs_suggest_apis` | Return compact API candidates for a task before using `godot-mcp-pro`. |

## Prerequisites

- Git
- Go matching the version in `go.mod`, if you build from source
- A local clone of the official Godot docs repository: <https://github.com/godotengine/godot-docs>

The repository also includes pre-built binaries:

| Platform | Architecture | MCP Server | Indexer |
|----------|--------------|------------|---------|
| Windows | amd64 | `godot-docs-mcp.exe` | `godot-docs-index.exe` |
| macOS | amd64 | `godot-docs-mcp-darwin-amd64` | `godot-docs-index-darwin-amd64` |
| macOS | arm64 | `godot-docs-mcp-darwin-arm64` | `godot-docs-index-darwin-arm64` |
| Linux | amd64 | `godot-docs-mcp-linux-amd64` | `godot-docs-index-linux-amd64` |
| Linux | arm64 | `godot-docs-mcp-linux-arm64` | `godot-docs-index-linux-arm64` |

## Quick Start: Windows

### 1. Clone this project

```powershell
cd E:\
git clone https://github.com/laishilong783-cmyk/Godot-docs-mcp.git
cd E:\godot-docs-mcp
```

Replace the repository URL with your actual GitHub repository URL.

### 2. Clone Godot docs

```powershell
cd E:\
git clone https://github.com/godotengine/godot-docs.git
cd E:\godot-docs
git checkout 4.4
```

Use another tag or branch if you want a different Godot documentation version.

### 3. Build the index

Using the included binary:

```powershell
cd E:\godot-docs-mcp
.\godot-docs-index.exe --docs-path E:\godot-docs --version 4.4 --db E:\godot_docs_4_4.db
```

Or build from source first:

```powershell
go build -o godot-docs-index.exe ./cmd/godot-docs-index
go build -o godot-docs-mcp.exe ./cmd/godot-docs-mcp
.\godot-docs-index.exe --docs-path E:\godot-docs --version 4.4 --db E:\godot_docs_4_4.db
```

### 4. Configure your MCP client

```json
{
  "mcpServers": {
    "godot-docs": {
      "command": "E:/godot-docs-mcp/godot-docs-mcp.exe",
      "args": [],
      "env": {
        "GODOT_DOCS_PATH": "E:/godot-docs",
        "GODOT_DOCS_DB": "E:/godot_docs_4_4.db",
        "GODOT_DOCS_VERSION": "4.4"
      }
    }
  }
}
```

## Quick Start: macOS

### 1. Clone this project

```bash
cd ~/dev
git clone https://github.com/laishilong783-cmyk/Godot-docs-mcp.git
cd ~/dev/godot-docs-mcp
```

Replace the repository URL with your actual GitHub repository URL.

### 2. Clone Godot docs

```bash
cd ~/dev
git clone https://github.com/godotengine/godot-docs.git
cd ~/dev/godot-docs
git checkout 4.4
```

### 3. Build the index

Apple Silicon:

```bash
cd ~/dev/godot-docs-mcp
chmod +x ./godot-docs-index-darwin-arm64 ./godot-docs-mcp-darwin-arm64
./godot-docs-index-darwin-arm64 --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

Intel Mac:

```bash
cd ~/dev/godot-docs-mcp
chmod +x ./godot-docs-index-darwin-amd64 ./godot-docs-mcp-darwin-amd64
./godot-docs-index-darwin-amd64 --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

Or build from source:

```bash
go build -o godot-docs-index ./cmd/godot-docs-index
go build -o godot-docs-mcp ./cmd/godot-docs-mcp
./godot-docs-index --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

### 4. Configure your MCP client

Apple Silicon example:

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

Use `godot-docs-mcp-darwin-amd64` on Intel Mac.

## Quick Start: Linux

### 1. Clone this project

```bash
cd ~/dev
git clone https://github.com/laishilong783-cmyk/Godot-docs-mcp.git
cd ~/dev/godot-docs-mcp
```

Replace the repository URL with your actual GitHub repository URL.

### 2. Clone Godot docs

```bash
cd ~/dev
git clone https://github.com/godotengine/godot-docs.git
cd ~/dev/godot-docs
git checkout 4.4
```

### 3. Build the index

Linux amd64:

```bash
cd ~/dev/godot-docs-mcp
chmod +x ./godot-docs-index-linux-amd64 ./godot-docs-mcp-linux-amd64
./godot-docs-index-linux-amd64 --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

Linux arm64:

```bash
cd ~/dev/godot-docs-mcp
chmod +x ./godot-docs-index-linux-arm64 ./godot-docs-mcp-linux-arm64
./godot-docs-index-linux-arm64 --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

Or build from source:

```bash
go build -o godot-docs-index ./cmd/godot-docs-index
go build -o godot-docs-mcp ./cmd/godot-docs-mcp
./godot-docs-index --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

### 4. Configure your MCP client

Linux amd64 example:

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

Use `godot-docs-mcp-linux-arm64` on Linux arm64.

## Multiple Godot Versions

You can store multiple documentation versions in one SQLite database. Index each docs checkout with a different `--version` value:

```bash
./godot-docs-index --docs-path ~/dev/godot-docs-4.3 --version 4.3 --db ~/dev/godot_docs_all.db
./godot-docs-index --docs-path ~/dev/godot-docs-4.4 --version 4.4 --db ~/dev/godot_docs_all.db
```

Set the default version in the MCP server:

```bash
export GODOT_DOCS_DB=~/dev/godot_docs_all.db
export GODOT_DOCS_VERSION=4.4
```

Each tool also accepts a `version` argument, so an agent can query a specific version:

```json
{"class_name": "CharacterBody2D", "method_name": "move_and_slide", "version": "4.4"}
```

## Example Queries

Search:

```json
{"query": "CharacterBody2D move_and_slide", "version": "4.4", "limit": 10}
```

Get class:

```json
{"class_name": "CharacterBody2D", "version": "4.4"}
```

Get filtered class members:

```json
{"class_name": "CharacterBody2D", "version": "4.4", "member_kinds": ["method"], "member_query": "slide", "member_limit": 10}
```

Get method:

```json
{"class_name": "CharacterBody2D", "method_name": "move_and_slide", "version": "4.4"}
```

Get property:

```json
{"class_name": "CharacterBody2D", "property_name": "velocity", "version": "4.4"}
```

Suggest APIs before operating Godot:

```json
{"task": "CharacterBody2D player movement velocity move_and_slide", "version": "4.4", "kinds": ["method", "property"], "limit": 12}
```

## Recommended Agent Workflow

1. Receive a Godot task.
2. Call `godot_docs_suggest_apis`, `godot_docs_search`, or `godot_docs_get_class`.
3. Confirm exact methods, properties, or signals with `godot_docs_get_method`, `godot_docs_get_property`, or `godot_docs_get_signal`.
4. Use `godot-mcp-pro` or another Godot operation tool to edit the project.
5. Run or inspect the project, then query docs again if errors appear.

## Common Errors

| Error | Fix |
|-------|-----|
| `GODOT_DOCS_PATH is not set` | Set `GODOT_DOCS_PATH` to your local `godot-docs` checkout. |
| `Index database not found` | Run `godot-docs-index` first and point `GODOT_DOCS_DB` to the generated database. |
| `Class not found` | Check the class name, selected `version`, and whether that version was indexed. |
| `Invalid path` | Only `.rst`, `.md`, and `.txt` files inside `GODOT_DOCS_PATH` are allowed. |

---

## 中文

[English](#godot-docs-mcp) | 中文

`godot-docs-mcp` 是一个本地离线 Godot 官方文档 MCP Server。它的作用是在 AI Agent 修改 Godot 项目前，先查询本地官方文档，确认真实存在的类、方法、属性和信号，再调用 `godot-mcp-pro` 等工具去操作 Godot。

这样可以减少以下问题：

- Agent 编造不存在的 Godot API。
- 混用 Godot 3.x 和 Godot 4.x 语法。
- 方法签名、属性名或信号参数错误。
- 直接操作 Godot 后生成错误节点、场景或脚本。

## 中文目录

- [功能特性](#功能特性)
- [MCP 工具](#mcp-工具)
- [前置要求](#前置要求)
- [Windows 使用步骤](#windows-使用步骤)
- [macOS 使用步骤](#macos-使用步骤)
- [Linux 使用步骤](#linux-使用步骤)
- [多版本 Godot 文档](#多版本-godot-文档)
- [查询示例](#查询示例)

## 功能特性

- 索引本地官方 `godot-docs` 仓库。
- 使用 SQLite 保存文档和 API 符号，并支持全文搜索。
- 支持在同一个数据库中保存多个 Godot 文档版本。
- 提供只读 MCP 工具，用于查询类、方法、属性、信号、文档页面和 API 建议。
- 文档克隆和索引完成后，可以离线使用。

## MCP 工具

| 工具 | 说明 |
|------|------|
| `godot_docs_search` | 全文搜索已索引文档。 |
| `godot_docs_get_page` | 按相对路径读取文档页面。 |
| `godot_docs_get_class` | 查询类信息，并可按条件返回成员。 |
| `godot_docs_get_method` | 查询方法签名和说明。 |
| `godot_docs_get_property` | 查询属性类型和说明。 |
| `godot_docs_get_signal` | 查询信号签名和说明。 |
| `godot_docs_suggest_apis` | 在调用 `godot-mcp-pro` 前，根据任务返回精简 API 候选。 |

## 前置要求

- Git
- 如果从源码构建，需要安装与 `go.mod` 匹配的 Go 版本
- 本地克隆官方 Godot 文档仓库：<https://github.com/godotengine/godot-docs>

本仓库也包含预编译二进制：

| 平台 | 架构 | MCP Server | 索引工具 |
|------|------|------------|----------|
| Windows | amd64 | `godot-docs-mcp.exe` | `godot-docs-index.exe` |
| macOS | amd64 | `godot-docs-mcp-darwin-amd64` | `godot-docs-index-darwin-amd64` |
| macOS | arm64 | `godot-docs-mcp-darwin-arm64` | `godot-docs-index-darwin-arm64` |
| Linux | amd64 | `godot-docs-mcp-linux-amd64` | `godot-docs-index-linux-amd64` |
| Linux | arm64 | `godot-docs-mcp-linux-arm64` | `godot-docs-index-linux-arm64` |

## Windows 使用步骤

### 1. 拉取本项目

```powershell
cd E:\
git clone https://github.com/laishilong783-cmyk/Godot-docs-mcp.git
cd E:\godot-docs-mcp
```

请把示例 URL 替换为你实际的 GitHub 仓库地址。

### 2. 拉取 Godot 官方文档

```powershell
cd E:\
git clone https://github.com/godotengine/godot-docs.git
cd E:\godot-docs
git checkout 4.4
```

如果你需要其他 Godot 版本，请 checkout 对应 tag 或分支。

### 3. 生成索引数据库

使用仓库内置 Windows 二进制：

```powershell
cd E:\godot-docs-mcp
.\godot-docs-index.exe --docs-path E:\godot-docs --version 4.4 --db E:\godot_docs_4_4.db
```

或者从源码构建：

```powershell
go build -o godot-docs-index.exe ./cmd/godot-docs-index
go build -o godot-docs-mcp.exe ./cmd/godot-docs-mcp
.\godot-docs-index.exe --docs-path E:\godot-docs --version 4.4 --db E:\godot_docs_4_4.db
```

### 4. 配置 MCP 客户端

```json
{
  "mcpServers": {
    "godot-docs": {
      "command": "E:/godot-docs-mcp/godot-docs-mcp.exe",
      "args": [],
      "env": {
        "GODOT_DOCS_PATH": "E:/godot-docs",
        "GODOT_DOCS_DB": "E:/godot_docs_4_4.db",
        "GODOT_DOCS_VERSION": "4.4"
      }
    }
  }
}
```

## macOS 使用步骤

### 1. 拉取本项目

```bash
cd ~/dev
git clone https://github.com/laishilong783-cmyk/Godot-docs-mcp.git
cd ~/dev/godot-docs-mcp
```

请把示例 URL 替换为你实际的 GitHub 仓库地址。

### 2. 拉取 Godot 官方文档

```bash
cd ~/dev
git clone https://github.com/godotengine/godot-docs.git
cd ~/dev/godot-docs
git checkout 4.4
```

### 3. 生成索引数据库

Apple Silicon:

```bash
cd ~/dev/godot-docs-mcp
chmod +x ./godot-docs-index-darwin-arm64 ./godot-docs-mcp-darwin-arm64
./godot-docs-index-darwin-arm64 --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

Intel Mac:

```bash
cd ~/dev/godot-docs-mcp
chmod +x ./godot-docs-index-darwin-amd64 ./godot-docs-mcp-darwin-amd64
./godot-docs-index-darwin-amd64 --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

或者从源码构建：

```bash
go build -o godot-docs-index ./cmd/godot-docs-index
go build -o godot-docs-mcp ./cmd/godot-docs-mcp
./godot-docs-index --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

### 4. 配置 MCP 客户端

Apple Silicon 示例：

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

Intel Mac 请把文件名换成 `godot-docs-mcp-darwin-amd64`。

## Linux 使用步骤

### 1. 拉取本项目

```bash
cd ~/dev
git clone https://github.com/laishilong783-cmyk/Godot-docs-mcp.git
cd ~/dev/godot-docs-mcp
```

请把示例 URL 替换为你实际的 GitHub 仓库地址。

### 2. 拉取 Godot 官方文档

```bash
cd ~/dev
git clone https://github.com/godotengine/godot-docs.git
cd ~/dev/godot-docs
git checkout 4.4
```

### 3. 生成索引数据库

Linux amd64:

```bash
cd ~/dev/godot-docs-mcp
chmod +x ./godot-docs-index-linux-amd64 ./godot-docs-mcp-linux-amd64
./godot-docs-index-linux-amd64 --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

Linux arm64:

```bash
cd ~/dev/godot-docs-mcp
chmod +x ./godot-docs-index-linux-arm64 ./godot-docs-mcp-linux-arm64
./godot-docs-index-linux-arm64 --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

或者从源码构建：

```bash
go build -o godot-docs-index ./cmd/godot-docs-index
go build -o godot-docs-mcp ./cmd/godot-docs-mcp
./godot-docs-index --docs-path ~/dev/godot-docs --version 4.4 --db ~/dev/godot_docs_4_4.db
```

### 4. 配置 MCP 客户端

Linux amd64 示例：

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

Linux arm64 请把文件名换成 `godot-docs-mcp-linux-arm64`。

## 多版本 Godot 文档

你可以把多个版本索引进同一个 SQLite 数据库。对每个文档目录使用不同的 `--version`：

```bash
./godot-docs-index --docs-path ~/dev/godot-docs-4.3 --version 4.3 --db ~/dev/godot_docs_all.db
./godot-docs-index --docs-path ~/dev/godot-docs-4.4 --version 4.4 --db ~/dev/godot_docs_all.db
```

设置 MCP Server 默认版本：

```bash
export GODOT_DOCS_DB=~/dev/godot_docs_all.db
export GODOT_DOCS_VERSION=4.4
```

每个工具也可以传入 `version` 参数，指定查询某个版本：

```json
{"class_name": "CharacterBody2D", "method_name": "move_and_slide", "version": "4.4"}
```

## 查询示例

搜索：

```json
{"query": "CharacterBody2D move_and_slide", "version": "4.4", "limit": 10}
```

查询类：

```json
{"class_name": "CharacterBody2D", "version": "4.4"}
```

按条件查询类成员：

```json
{"class_name": "CharacterBody2D", "version": "4.4", "member_kinds": ["method"], "member_query": "slide", "member_limit": 10}
```

查询方法：

```json
{"class_name": "CharacterBody2D", "method_name": "move_and_slide", "version": "4.4"}
```

查询属性：

```json
{"class_name": "CharacterBody2D", "property_name": "velocity", "version": "4.4"}
```

在操作 Godot 前获取 API 建议：

```json
{"task": "CharacterBody2D player movement velocity move_and_slide", "version": "4.4", "kinds": ["method", "property"], "limit": 12}
```

## 推荐 Agent 工作流

1. Agent 收到 Godot 任务。
2. 先调用 `godot_docs_suggest_apis`、`godot_docs_search` 或 `godot_docs_get_class`。
3. 用 `godot_docs_get_method`、`godot_docs_get_property`、`godot_docs_get_signal` 精确确认 API。
4. 再调用 `godot-mcp-pro` 或其他 Godot 操作工具修改项目。
5. 运行或检查项目，如果报错，再回查文档。

## 常见错误

| 错误 | 处理方式 |
|------|----------|
| `GODOT_DOCS_PATH is not set` | 把 `GODOT_DOCS_PATH` 设置为本地 `godot-docs` 目录。 |
| `Index database not found` | 先运行 `godot-docs-index`，并让 `GODOT_DOCS_DB` 指向生成的数据库。 |
| `Class not found` | 检查类名、查询版本，以及该版本是否已经索引。 |
| `Invalid path` | 只能读取 `GODOT_DOCS_PATH` 内的 `.rst`、`.md`、`.txt` 文件。 |
