package cmd

import "testing"

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

func TestParseMatrixTSV(t *testing.T) {
	t.Parallel()

	matrix, err := parseMatrix("A\tB\nC\tD\n")
	if err != nil {
		t.Fatalf("parseMatrix returned error: %v", err)
	}

	if got, want := matrix[1][0], "C"; got != want {
		t.Fatalf("unexpected value: got %q want %q", got, want)
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
	if values[2] != "" || values[3] != "" {
		t.Fatalf("expected padded empty values, got %#v", values)
	}
}
