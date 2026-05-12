package docs

import (
	"strings"
	"testing"
)

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"CharacterBody2D\n===============\n\nSome text", "CharacterBody2D"},
		{"Using CharacterBody2D\n=====================\n\nSome text", "Using CharacterBody2D"},
		{"\n\nSome title\n----------", "Some title"},
	}

	for _, tt := range tests {
		got := extractTitle(tt.input)
		if got != tt.want {
			t.Errorf("extractTitle(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestClassFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"classes/class_characterbody2d.rst", "Characterbody2D"},
		{"classes/class_node2d.rst", "Node2D"},
		{"classes/class_animationtree.rst", "Animationtree"},
	}

	for _, tt := range tests {
		got := ClassFromPath(tt.path)
		if got != tt.want {
			t.Errorf("ClassFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestSnakeToPascal(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"character_body_2d", "CharacterBody2D"},
		{"node_2d", "Node2D"},
		{"animation_tree", "AnimationTree"},
		{"@gdscript", "@Gdscript"},
	}

	for _, tt := range tests {
		got := snakeToPascal(tt.input)
		if got != tt.want {
			t.Errorf("snakeToPascal(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCleanInlineRST(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"This is ``code`` here", "This is code here"},
		{"*emphasis* and **strong**", "emphasis and strong"},
		{"See :ref:`Node` class", "See Node class"},
		{"Link to `page`_ here", "Link to page_ here"},
	}

	for _, tt := range tests {
		got := cleanInlineRST(tt.input)
		if got != tt.want {
			t.Errorf("cleanInlineRST(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsRSTUnderline(t *testing.T) {
	if !isRSTUnderline("===") {
		t.Error("Expected === to be underline")
	}
	if !isRSTUnderline("-----") {
		t.Error("Expected ----- to be underline")
	}
	if isRSTUnderline("abc") {
		t.Error("Expected abc to NOT be underline")
	}
	if isRSTUnderline("==-") {
		t.Error("Expected ==- to NOT be underline")
	}
}

func TestParseMembersIgnoresPropertySetgetBlocks(t *testing.T) {
	content := "\n" +
		".. rst-class:: classref-property\n\n" +
		":ref:`Vector2<class_Vector2>` **velocity** = ``Vector2(0, 0)``\n\n" +
		".. rst-class:: classref-property-setget\n\n" +
		"- |void| **set_velocity**\\ (\\ value\\: :ref:`Vector2<class_Vector2>`\\ )\n" +
		"- :ref:`Vector2<class_Vector2>` **get_velocity**\\ (\\ )\n\n" +
		"Current velocity vector in pixels per second.\n\n" +
		".. rst-class:: classref-item-separator\n\n" +
		"----\n\n" +
		".. rst-class:: classref-method\n\n" +
		":ref:`bool<class_bool>` **move_and_slide**\\ (\\ )\n\n" +
		"Moves the body based on velocity.\n"

	symbols := parseMembersFromRawRST(content, "CharacterBody2D", "classes/class_characterbody2d.rst")

	var properties, methods []Symbol
	for _, sym := range symbols {
		switch sym.Kind {
		case KindProperty:
			properties = append(properties, sym)
		case KindMethod:
			methods = append(methods, sym)
		}
		if sym.MemberName == "set_velocity" || sym.MemberName == "get_velocity" {
			t.Fatalf("property set/get helper was indexed as a member: %+v", sym)
		}
	}

	if len(properties) != 1 || properties[0].MemberName != "velocity" {
		t.Fatalf("properties = %+v, want only velocity", properties)
	}
	if strings.Contains(properties[0].Description, "set_velocity") || strings.Contains(properties[0].Description, "get_velocity") {
		t.Fatalf("property description contains set/get helper rows: %q", properties[0].Description)
	}
	if len(methods) != 1 || methods[0].MemberName != "move_and_slide" {
		t.Fatalf("methods = %+v, want only move_and_slide", methods)
	}
}
