package cmd

import (
	"reflect"
	"testing"
)

func TestNormalizeCLIArgsLeavesRegularArgsUntouched(t *testing.T) {
	got := normalizeCLIArgs([]string{"pr", "list", "--limit", "5"})
	want := []string{"pr", "list", "--limit", "5"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestNormalizeCLIArgsExpandsBareJSONFlag(t *testing.T) {
	got := normalizeCLIArgs([]string{"pr", "list", "--json"})
	want := []string{"pr", "list", "--json=*"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestNormalizeCLIArgsConsumesJSONValue(t *testing.T) {
	got := normalizeCLIArgs([]string{"pr", "view", "1", "--json", "id,title"})
	want := []string{"pr", "view", "1", "--json=id,title"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestNormalizeCLIArgsExpandsAliasBeforeJSONNormalization(t *testing.T) {
	got := normalizeCLIArgsWithAliases([]string{"pv", "7", "--json"}, map[string]string{
		"pv": "pr view",
	})
	want := []string{"pr", "view", "7", "--json=*"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestExpandAliasArgsStopsOnAliasLoop(t *testing.T) {
	got := expandAliasArgs([]string{"a"}, map[string]string{
		"a": "b",
		"b": "a",
	})
	want := []string{"a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
