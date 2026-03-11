package cmd

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestRunExtensionCommand(t *testing.T) {
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
