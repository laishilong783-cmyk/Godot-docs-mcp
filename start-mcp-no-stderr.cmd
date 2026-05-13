@echo off
chcp 65001 >nul
set PYTHONIOENCODING=utf-8
if not defined GODOT_DOCS_PATH set GODOT_DOCS_PATH=E:\godot\godoc-docs
if not defined GODOT_DOCS_DB set GODOT_DOCS_DB=E:\godot\godot_docs_4_4.db
if not defined GODOT_DOCS_VERSION set GODOT_DOCS_VERSION=4.4
E:\godot\godot-docs-mcp\godot-docs-mcp.exe 2>nul
