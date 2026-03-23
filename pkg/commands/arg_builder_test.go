package commands

import (
	"reflect"
	"testing"
)

func TestArgBuilder_NewWithInitialArgs(t *testing.T) {
	b := NewArgBuilder("search", "query")
	got := b.Build()
	want := []string{"search", "query"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build() = %v, want %v", got, want)
	}
}

func TestArgBuilder_NewEmpty(t *testing.T) {
	b := NewArgBuilder()
	got := b.Build()
	if len(got) != 0 {
		t.Errorf("Build() = %v, want empty slice", got)
	}
}

func TestArgBuilder_Add(t *testing.T) {
	got := NewArgBuilder("cmd").Add("--flag", "value").Build()
	want := []string{"cmd", "--flag", "value"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build() = %v, want %v", got, want)
	}
}

func TestArgBuilder_AddIf_True(t *testing.T) {
	got := NewArgBuilder("cmd").AddIf(true, "--verbose").Build()
	want := []string{"cmd", "--verbose"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build() = %v, want %v", got, want)
	}
}

func TestArgBuilder_AddIf_False(t *testing.T) {
	got := NewArgBuilder("cmd").AddIf(false, "--verbose").Build()
	want := []string{"cmd"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build() = %v, want %v", got, want)
	}
}

func TestArgBuilder_AddIf_WithValue(t *testing.T) {
	got := NewArgBuilder("cmd").AddIf(true, "--sort", "created").Build()
	want := []string{"cmd", "--sort", "created"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build() = %v, want %v", got, want)
	}
}

func TestArgBuilder_Chaining(t *testing.T) {
	got := NewArgBuilder("search").
		Add("query").
		AddIf(true, "-s", "created").
		AddIf(false, "--everything").
		AddIf(true, "--content").
		Build()

	want := []string{"search", "query", "-s", "created", "--content"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build() = %v, want %v", got, want)
	}
}
