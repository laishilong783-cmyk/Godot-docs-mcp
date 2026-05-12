package docs

import (
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
