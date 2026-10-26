package apicodegen

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := FindRepoRoot(".")
	if err != nil {
		t.Fatalf("find repository root: %v", err)
	}
	return root
}

func TestParseCodesReadsTheServerCodes(t *testing.T) {
	root := repoRoot(t)

	codes, err := ParseCodes(filepath.Join(root, CodesSource))
	if err != nil {
		t.Fatalf("parse codes: %v", err)
	}

	if len(codes) == 0 {
		t.Fatal("no codes were parsed")
	}

	if !slices.IsSorted(codes) {
		t.Errorf("codes are not sorted: %v", codes)
	}

	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		if seen[code] {
			t.Errorf("code %q appears more than once", code)
		}
		seen[code] = true

		if !isSnakeCase(code) {
			t.Errorf("code %q is not lower snake case", code)
		}
	}
}

func isSnakeCase(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '_':
		default:
			return false
		}
	}
	return true
}

func TestParseCodesIgnoresNonCodeConstants(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "codes.go")
	contents := `package response

const (
	CodeAlpha = "alpha"
	CodeBeta  = "beta"
	notExported = "hidden"
	Other     = "other"
	Count     = 3
)

const CodeGamma = "gamma"
`
	if err := os.WriteFile(source, []byte(contents), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	codes, err := ParseCodes(source)
	if err != nil {
		t.Fatalf("parse codes: %v", err)
	}

	want := []string{"alpha", "beta", "gamma"}
	if !slices.Equal(codes, want) {
		t.Errorf("codes = %v, want %v", codes, want)
	}
}

func TestParseCodesRejectsAFileWithoutCodes(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "codes.go")
	if err := os.WriteFile(source, []byte("package response\n\nconst Nothing = 1\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if _, err := ParseCodes(source); err == nil {
		t.Fatal("expected an error for a file with no codes")
	}
}

func TestRenderProducesAUsableModule(t *testing.T) {
	module := Render([]string{"alpha", "beta"})

	for _, want := range []string{
		"export const API_CODES = [",
		"'alpha',",
		"'beta',",
		"export type ApiCode = (typeof API_CODES)[number];",
		"export function isApiCode(value: string): value is ApiCode {",
	} {
		if !strings.Contains(module, want) {
			t.Errorf("module is missing %q", want)
		}
	}
	if !strings.Contains(module, "DO NOT EDIT") {
		t.Error("module should say it is generated")
	}
}

func TestGeneratedClientIsUpToDate(t *testing.T) {
	root := repoRoot(t)

	want, err := Generate(root)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	path := filepath.Join(root, OutputPath)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", OutputPath, err)
	}

	// Git rewrites line endings on checkout, so a Windows working tree holds
	// CRLF where the generator writes LF. Compare the content, not the bytes.
	if !sameContent(string(got), want) {
		t.Errorf("%s is out of date; run `task.ps1 generate`", OutputPath)
	}
}

// sameContent compares two files, ignoring their line endings.
func sameContent(a, b string) bool {
	normalise := func(value string) string {
		return strings.ReplaceAll(value, "\r\n", "\n")
	}
	return normalise(a) == normalise(b)
}
