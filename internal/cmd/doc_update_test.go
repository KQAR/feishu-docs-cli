package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestDocCommandStructure(t *testing.T) {
	t.Parallel()

	docCmd := newDocCmd()
	got := subcommandNames(docCmd)
	want := []string{"create", "get", "content", "blocks", "block", "update"}

	assertCommandNames(t, got, want)
}

func TestDocUpdateCommandStructure(t *testing.T) {
	t.Parallel()

	updateCmd := newDocUpdateCmd()
	got := subcommandNames(updateCmd)
	want := []string{"append", "insert", "set-text", "delete", "table"}

	assertCommandNames(t, got, want)
}

func TestDocUpdateTableCommandStructure(t *testing.T) {
	t.Parallel()

	tableCmd := newDocUpdateTableCmd()
	got := subcommandNames(tableCmd)
	want := []string{
		"create",
		"show",
		"write",
		"set-cell",
		"insert-row",
		"insert-column",
		"delete-rows",
		"delete-columns",
		"merge",
		"unmerge",
		"props",
	}

	assertCommandNames(t, got, want)
}

func TestParseLineToBlockStripsMarkdownPrefixes(t *testing.T) {
	t.Parallel()

	bullet := parseLineToBlock("- item")
	if got := textElementsPlain(bullet.Bullet.Elements); got != "item" {
		t.Fatalf("unexpected bullet content: %q", got)
	}

	heading := parseLineToBlock("## title")
	if got := textElementsPlain(heading.Heading2.Elements); got != "title" {
		t.Fatalf("unexpected heading content: %q", got)
	}

	ordered := parseLineToBlock("12. ordered item")
	if got := textElementsPlain(ordered.Ordered.Elements); got != "ordered item" {
		t.Fatalf("unexpected ordered content: %q", got)
	}
}

func TestParseMarkdownToBlocksCodeFence(t *testing.T) {
	t.Parallel()

	md := "```\nfoo\nbar\n```"
	blocks := parseMarkdownToBlocks(md)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 code block, got %d", len(blocks))
	}
	if blocks[0].Code == nil {
		t.Fatal("expected code block")
	}
	got := textElementsPlain(blocks[0].Code.Elements)
	if got != "foo\nbar" {
		t.Fatalf("unexpected code content: %q", got)
	}
}

func TestParseMarkdownToBlocksHorizontalRule(t *testing.T) {
	t.Parallel()

	for _, hr := range []string{"---", "***", "___", "- - -"} {
		blocks := parseMarkdownToBlocks(hr)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block for %q, got %d", hr, len(blocks))
		}
		if *blocks[0].BlockType != 22 {
			t.Fatalf("expected divider block type 22 for %q, got %d", hr, *blocks[0].BlockType)
		}
	}
}

func TestParseMarkdownToBlocksBlockquote(t *testing.T) {
	t.Parallel()

	md := "> hello\n> world"
	blocks := parseMarkdownToBlocks(md)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks from blockquote, got %d", len(blocks))
	}
	if got := textElementsPlain(blocks[0].Text.Elements); got != "hello" {
		t.Fatalf("unexpected first quote line: %q", got)
	}
}

func TestIsHorizontalRule(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		input string
		want  bool
	}{
		{"---", true},
		{"***", true},
		{"___", true},
		{"- - -", true},
		{"----", true},
		{"--", false},
		{"-*-", false},
		{"text", false},
	} {
		if got := isHorizontalRule(tc.input); got != tc.want {
			t.Errorf("isHorizontalRule(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestResolveDocumentIDPassthrough(t *testing.T) {
	t.Parallel()

	// Plain doc IDs should pass through unchanged
	if got := resolveDocumentID("doxcnAbCdEf"); got != "doxcnAbCdEf" {
		t.Fatalf("expected passthrough, got %q", got)
	}
}

func subcommandNames(cmd interface{ Commands() []*cobra.Command }) []string {
	commands := cmd.Commands()
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.Name())
	}
	return names
}

func assertCommandNames(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("unexpected command count: got %v want %v", got, want)
	}

	expected := make(map[string]struct{}, len(want))
	for _, name := range want {
		expected[name] = struct{}{}
	}
	for _, name := range got {
		if _, ok := expected[name]; !ok {
			t.Fatalf("unexpected command set: got %v want %v", got, want)
		}
	}
}
