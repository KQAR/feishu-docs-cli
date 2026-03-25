package cmd

import (
	"testing"
)

func TestNormalizeMatrixInferAndPad(t *testing.T) {
	t.Parallel()

	matrix, err := normalizeMatrix([][]string{
		{"H1", "H2"},
		{"A1"},
	}, 0, 0)
	if err != nil {
		t.Fatalf("normalizeMatrix returned error: %v", err)
	}

	if len(matrix) != 2 || len(matrix[0]) != 2 {
		t.Fatalf("unexpected matrix shape: %#v", matrix)
	}
	if matrix[1][1] != "" {
		t.Fatalf("expected padded empty cell, got %q", matrix[1][1])
	}
}

func TestNormalizeMatrixRejectOversizedInput(t *testing.T) {
	t.Parallel()

	_, err := normalizeMatrix([][]string{{"A", "B", "C"}}, 1, 2)
	if err == nil {
		t.Fatal("expected error for oversized input")
	}
}

func TestNormalizeMatrixRejectOversizedRows(t *testing.T) {
	t.Parallel()

	_, err := normalizeMatrix([][]string{{"A"}, {"B"}, {"C"}}, 2, 1)
	if err == nil {
		t.Fatal("expected error for oversized row count")
	}
}

func TestNormalizeMatrixRejectNegative(t *testing.T) {
	t.Parallel()

	_, err := normalizeMatrix([][]string{{"A"}}, -1, 1)
	if err == nil {
		t.Fatal("expected error for negative rows")
	}

	_, err = normalizeMatrix([][]string{{"A"}}, 1, -1)
	if err == nil {
		t.Fatal("expected error for negative cols")
	}
}

func TestNormalizeMatrixEmptyRequiresExplicitSize(t *testing.T) {
	t.Parallel()

	_, err := normalizeMatrix([][]string{}, 0, 0)
	if err == nil {
		t.Fatal("expected error for empty matrix with no explicit size")
	}
}

func TestNormalizeMatrixExplicitSizePads(t *testing.T) {
	t.Parallel()

	matrix, err := normalizeMatrix([][]string{{"A"}}, 3, 2)
	if err != nil {
		t.Fatalf("normalizeMatrix returned error: %v", err)
	}

	if len(matrix) != 3 || len(matrix[0]) != 2 {
		t.Fatalf("unexpected shape: %d x %d", len(matrix), len(matrix[0]))
	}
	if matrix[0][0] != "A" || matrix[0][1] != "" || matrix[1][0] != "" || matrix[2][1] != "" {
		t.Fatalf("unexpected values: %#v", matrix)
	}
}

func TestParseMatrixTSV(t *testing.T) {
	t.Parallel()

	matrix, err := parseMatrix("A\tB\nC\tD\n")
	if err != nil {
		t.Fatalf("parseMatrix returned error: %v", err)
	}

	if len(matrix) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(matrix))
	}
	if got, want := matrix[0][0], "A"; got != want {
		t.Fatalf("unexpected value: got %q want %q", got, want)
	}
	if got, want := matrix[1][0], "C"; got != want {
		t.Fatalf("unexpected value: got %q want %q", got, want)
	}
	if got, want := matrix[1][1], "D"; got != want {
		t.Fatalf("unexpected value: got %q want %q", got, want)
	}
}

func TestParseMatrixJSON(t *testing.T) {
	t.Parallel()

	matrix, err := parseMatrix(`[["X","Y"],["1","2"]]`)
	if err != nil {
		t.Fatalf("parseMatrix returned error: %v", err)
	}

	if len(matrix) != 2 || len(matrix[0]) != 2 {
		t.Fatalf("unexpected shape: %#v", matrix)
	}
	if matrix[0][0] != "X" || matrix[1][1] != "2" {
		t.Fatalf("unexpected values: %#v", matrix)
	}
}

func TestParseMatrixJSONInvalid(t *testing.T) {
	t.Parallel()

	_, err := parseMatrix(`[["X","Y"`)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseMatrixEmpty(t *testing.T) {
	t.Parallel()

	matrix, err := parseMatrix("")
	if err != nil {
		t.Fatalf("parseMatrix returned error: %v", err)
	}
	if len(matrix) != 0 {
		t.Fatalf("expected empty matrix, got %#v", matrix)
	}
}

func TestParseMatrixWindowsLineEndings(t *testing.T) {
	t.Parallel()

	matrix, err := parseMatrix("A\tB\r\nC\tD\r\n")
	if err != nil {
		t.Fatalf("parseMatrix returned error: %v", err)
	}

	if len(matrix) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(matrix))
	}
	if matrix[0][1] != "B" || matrix[1][0] != "C" {
		t.Fatalf("unexpected values with CRLF: %#v", matrix)
	}
}

func TestParseVectorInputPadsToSize(t *testing.T) {
	t.Parallel()

	values, err := parseVectorInput(`["A","B"]`, "", 4, "\t")
	if err != nil {
		t.Fatalf("parseVectorInput returned error: %v", err)
	}

	if len(values) != 4 {
		t.Fatalf("unexpected vector length: %d", len(values))
	}
	if values[0] != "A" || values[1] != "B" {
		t.Fatalf("unexpected values: %#v", values)
	}
	if values[2] != "" || values[3] != "" {
		t.Fatalf("expected padded empty values, got %#v", values)
	}
}

func TestParseVectorInputRejectOversized(t *testing.T) {
	t.Parallel()

	_, err := parseVectorInput(`["A","B","C"]`, "", 2, "\t")
	if err == nil {
		t.Fatal("expected error for oversized vector")
	}
}

func TestParseVectorInputTabSeparated(t *testing.T) {
	t.Parallel()

	values, err := parseVectorInput("X\tY\tZ", "", 3, "\t")
	if err != nil {
		t.Fatalf("parseVectorInput returned error: %v", err)
	}

	if len(values) != 3 {
		t.Fatalf("unexpected length: %d", len(values))
	}
	if values[0] != "X" || values[1] != "Y" || values[2] != "Z" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestParseVectorInputNewlineSeparated(t *testing.T) {
	t.Parallel()

	values, err := parseVectorInput("A\nB\nC", "", 3, "\n")
	if err != nil {
		t.Fatalf("parseVectorInput returned error: %v", err)
	}

	if values[0] != "A" || values[1] != "B" || values[2] != "C" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestParseVectorInputEmpty(t *testing.T) {
	t.Parallel()

	values, err := parseVectorInput("", "", 3, "\t")
	if err != nil {
		t.Fatalf("parseVectorInput returned error: %v", err)
	}

	if len(values) != 3 {
		t.Fatalf("unexpected length: %d", len(values))
	}
	for i, v := range values {
		if v != "" {
			t.Fatalf("expected empty string at index %d, got %q", i, v)
		}
	}
}

func TestReadOptionalInputMutualExclusion(t *testing.T) {
	t.Parallel()

	_, err := readOptionalInput("some data", "some-file.txt")
	if err == nil {
		t.Fatal("expected error when both data and dataFile provided")
	}
}

func TestParseColumnWidths(t *testing.T) {
	t.Parallel()

	widths, err := parseColumnWidths("100,200,150", 3)
	if err != nil {
		t.Fatalf("parseColumnWidths returned error: %v", err)
	}

	if len(widths) != 3 || widths[0] != 100 || widths[1] != 200 || widths[2] != 150 {
		t.Fatalf("unexpected widths: %#v", widths)
	}
}

func TestParseColumnWidthsEmpty(t *testing.T) {
	t.Parallel()

	widths, err := parseColumnWidths("", 3)
	if err != nil {
		t.Fatalf("parseColumnWidths returned error: %v", err)
	}
	if widths != nil {
		t.Fatalf("expected nil widths, got %#v", widths)
	}
}

func TestParseColumnWidthsMismatch(t *testing.T) {
	t.Parallel()

	_, err := parseColumnWidths("100,200", 3)
	if err == nil {
		t.Fatal("expected error for width count mismatch")
	}
}

func TestParseColumnWidthsInvalid(t *testing.T) {
	t.Parallel()

	_, err := parseColumnWidths("100,abc,200", 3)
	if err == nil {
		t.Fatal("expected error for invalid width value")
	}
}

func TestNewTemporaryBlockID(t *testing.T) {
	t.Parallel()

	id := newTemporaryBlockID("cell_0_1")
	if id != "tmp_cell_0_1" {
		t.Fatalf("unexpected ID: %s", id)
	}

	id = newTemporaryBlockID("cell text")
	if id != "tmp_cell_text" {
		t.Fatalf("unexpected ID with space: %s", id)
	}

	id = newTemporaryBlockID("my-block")
	if id != "tmp_my_block" {
		t.Fatalf("unexpected ID with dash: %s", id)
	}
}

func TestTableSnapshotCellID(t *testing.T) {
	t.Parallel()

	s := &tableSnapshot{
		Rows:    2,
		Cols:    3,
		CellIDs: []string{"a", "b", "c", "d", "e", "f"},
	}

	id, err := s.cellID(0, 0)
	if err != nil || id != "a" {
		t.Fatalf("(0,0) expected 'a', got %q err %v", id, err)
	}

	id, err = s.cellID(1, 2)
	if err != nil || id != "f" {
		t.Fatalf("(1,2) expected 'f', got %q err %v", id, err)
	}

	_, err = s.cellID(-1, 0)
	if err == nil {
		t.Fatal("expected error for negative row")
	}

	_, err = s.cellID(0, 3)
	if err == nil {
		t.Fatal("expected error for out-of-bounds col")
	}

	_, err = s.cellID(2, 0)
	if err == nil {
		t.Fatal("expected error for out-of-bounds row")
	}
}

func TestTableSnapshotCellIDInsufficientCells(t *testing.T) {
	t.Parallel()

	s := &tableSnapshot{
		Rows:    2,
		Cols:    2,
		CellIDs: []string{"a", "b"}, // only 2, need 4
	}

	_, err := s.cellID(1, 0)
	if err == nil {
		t.Fatal("expected error for insufficient cells")
	}
}

func TestBuildTableDescendants(t *testing.T) {
	t.Parallel()

	matrix := [][]string{
		{"H1", "H2"},
		{"A", "B"},
	}

	tableID, descendants := buildTableDescendants("parent1", matrix, nil, true, false)
	if tableID != "tmp_table" {
		t.Fatalf("unexpected table ID: %s", tableID)
	}

	// 1 table + 4 cells + 4 text blocks = 9
	if len(descendants) != 9 {
		t.Fatalf("expected 9 descendants, got %d", len(descendants))
	}

	// First descendant should be the table block
	if descendants[0].BlockType == nil || *descendants[0].BlockType != docxBlockTypeTable {
		t.Fatalf("first descendant should be a table block")
	}
}

func TestBuildTableDescendantsEmptyCells(t *testing.T) {
	t.Parallel()

	matrix := [][]string{
		{"A", ""},
		{"", "B"},
	}

	_, descendants := buildTableDescendants("parent1", matrix, nil, false, false)
	// 1 table + 4 cells + 2 text blocks (only non-empty cells get text) = 7
	if len(descendants) != 7 {
		t.Fatalf("expected 7 descendants, got %d", len(descendants))
	}
}

func TestBuildTableDescendantsWithColumnWidths(t *testing.T) {
	t.Parallel()

	matrix := [][]string{{"A", "B"}}
	widths := []int{100, 200}

	_, descendants := buildTableDescendants("p", matrix, widths, false, false)
	if descendants[0].Table == nil {
		t.Fatal("table block missing table property")
	}
}

func TestBlockPlainText(t *testing.T) {
	t.Parallel()

	if text := blockPlainText(nil); text != "" {
		t.Fatalf("expected empty for nil block, got %q", text)
	}
}

func TestTextElementsPlainEmpty(t *testing.T) {
	t.Parallel()

	if text := textElementsPlain(nil); text != "" {
		t.Fatalf("expected empty for nil elements, got %q", text)
	}
}
