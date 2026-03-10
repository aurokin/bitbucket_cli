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
