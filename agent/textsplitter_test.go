package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/thomas-marquis/goLLMan/pkg"
)

func TestOnlyContainsHeaders_Snapshots(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace only", "   \n\t\n"},

		// ATX headings
		{"single atx", "# Title"},
		{"multiple atx with blanks", "# H1\n\n## H2\n\n### H3"},
		{"atx with up to 3 leading spaces", "   ## Indented"},
		{"atx with 4 leading spaces (not a heading)", "    # Not Heading"},
		{"atx with trailing hashes", "# Title ####"},
		{"atx no-space after hashes (not a heading)", "##Heading"},
		{"atx with trailing spaces", "## Title   "},

		// Setext headings
		{"setext dashed", "Title\n---"},
		{"setext equals", "Another Title\n====="},
		{"setext with blank line between (accepted by function)", "Title\n\n---"},
		{"setext underline alone (not a heading)", "-----"},

		// Mixed content
		{"heading then paragraph", "# H1\nParagraph text"},
		{"paragraph then heading", "Text first\n# H1"},
		{"code fence (not heading)", "```\ncode\n```"},
		{"list item (not heading)", "- item"},

		// Newline variations
		{"setext with CRLF", "Title\r\n---"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := OnlyContainsHeaders(tc.in)
			snaps.MatchSnapshot(t, got)
		})
	}
}

func TestOnlyContainsHeaders_TestdataFiles_Snapshot(t *testing.T) {
	// List files explicitly to ensure deterministic order for snapshots.
	files := []string{
		"only_headers.md",
		"setext_only.md",
		"book_like.md",
		"tricky.md",
	}

	results := make(map[string]bool, len(files))
	for _, name := range files {
		path := filepath.Join("testdata", name)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		results[name] = OnlyContainsHeaders(string(b))
	}

	snaps.MatchSnapshot(t, results)
}

func TestSplitDocuments_Snapshot(t *testing.T) {
	// Prepare splitter
	splitter := makeTextSplitter()

	// Define input files for splitting
	files := []string{
		"book_like.md",
		"tricky.md",
	}

	// Build input documents
	var inputs []*ai.Document
	for _, name := range files {
		path := filepath.Join("testdata", name)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		inputs = append(inputs, ai.DocumentFromText(string(b), map[string]any{
			"source": name,
		}))
	}

	// Execute split
	out, err := SplitDocuments(splitter, inputs)
	if err != nil {
		t.Fatalf("SplitDocuments error: %v", err)
	}

	// Simplify to snapshot-friendly structure
	type SnapDoc struct {
		Text     string         `json:"text"`
		Metadata map[string]any `json:"metadata"`
	}
	simplified := make([]SnapDoc, 0, len(out))
	for _, d := range out {
		simplified = append(simplified, SnapDoc{
			Text:     pkg.ContentToText(d.Content),
			Metadata: d.Metadata,
		})
	}

	snaps.MatchSnapshot(t, simplified)
}
