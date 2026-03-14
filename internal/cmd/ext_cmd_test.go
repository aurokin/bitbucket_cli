package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestListAliasEntriesSorted(t *testing.T) {
	t.Parallel()

	got := listAliasEntries(map[string]string{
		"z": "pr view",
		"a": "pr list",
	})
	want := []aliasEntry{
		{Name: "a", Expansion: "pr list"},
		{Name: "z", Expansion: "pr view"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestDiscoverExtensions(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)

	executable := filepath.Join(dir, "bb-hello")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write extension: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bb-data.txt"), []byte("not executable"), 0o644); err != nil {
		t.Fatalf("write non-extension: %v", err)
	}

	extensions, err := discoverExtensions()
	if err != nil {
		t.Fatalf("discoverExtensions returned error: %v", err)
	}

	want := []extensionEntry{{Name: "hello", Executable: executable}}
	if !reflect.DeepEqual(extensions, want) {
		t.Fatalf("expected %+v, got %+v", want, extensions)
	}
}

func TestExtensionEntriesSorted(t *testing.T) {
	t.Parallel()

	got := extensionEntries(map[string]string{
		"zeta":  "/tmp/bb-zeta",
		"alpha": "/tmp/bb-alpha",
	})
	want := []extensionEntry{
		{Name: "alpha", Executable: "/tmp/bb-alpha"},
		{Name: "zeta", Executable: "/tmp/bb-zeta"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestIsExtensionCandidate(t *testing.T) {
	t.Parallel()

	if !isExtensionCandidate("bb-hello") {
		t.Fatal("expected bb-hello to be a candidate")
	}
	for _, name := range []string{"bb", "hello", ""} {
		if isExtensionCandidate(name) {
			t.Fatalf("did not expect %q to be a candidate", name)
		}
	}
}

func TestRunExtensionCommand(t *testing.T) {
	lockCommandTestHooks(t)

	originalLookPath := execLookPath
	originalExecute := executeExternalFunc
	t.Cleanup(func() {
		execLookPath = originalLookPath
		executeExternalFunc = originalExecute
	})

	var gotExecutable string
	var gotArgs []string
	execLookPath = func(file string) (string, error) {
		return "/tmp/" + file, nil
	}
	executeExternalFunc = func(executable string, args []string) error {
		gotExecutable = executable
		gotArgs = append([]string(nil), args...)
		return nil
	}

	if err := runExtensionCommand("hello", []string{"one", "two"}); err != nil {
		t.Fatalf("runExtensionCommand returned error: %v", err)
	}

	if gotExecutable != "/tmp/bb-hello" {
		t.Fatalf("unexpected executable %q", gotExecutable)
	}
	if !reflect.DeepEqual(gotArgs, []string{"one", "two"}) {
		t.Fatalf("unexpected args %v", gotArgs)
	}
}

func TestAliasSetCommandOutput(t *testing.T) {
	t.Setenv("BB_CONFIG_DIR", t.TempDir())

	output := renderCommand(t, "alias", "set", "pv", "pr", "view")
	for _, expected := range []string{
		"Set alias pv=pr view",
		"Next: bb alias get pv",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output, got %q", expected, output)
		}
	}
	assertOrderedSubstrings(t, output,
		"Set alias pv=pr view",
		"Next: bb alias get pv",
	)
}

func TestAliasDeleteCommandOutput(t *testing.T) {
	t.Setenv("BB_CONFIG_DIR", t.TempDir())
	_ = renderCommand(t, "alias", "set", "pv", "pr", "view")

	output := renderCommand(t, "alias", "delete", "pv")
	for _, expected := range []string{
		"Deleted alias pv",
		"Next: bb alias list",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output, got %q", expected, output)
		}
	}
	assertOrderedSubstrings(t, output,
		"Deleted alias pv",
		"Next: bb alias list",
	)
}

func TestAliasListEmptyOutput(t *testing.T) {
	t.Setenv("BB_CONFIG_DIR", t.TempDir())

	output := renderCommand(t, "alias", "list")
	if !strings.Contains(output, "No aliases configured.") {
		t.Fatalf("expected empty alias output, got %q", output)
	}
}
