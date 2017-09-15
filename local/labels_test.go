package local

import (
	"reflect"
	"testing"
)

func TestParseLabels(t *testing.T) {
	set, unSet := ParseLabels("foo,!bar,baz")

	expectedSet := map[string]bool{"foo": true, "baz": true}
	expectedUnSet := map[string]bool{"bar": true}

	if !reflect.DeepEqual(expectedSet, set) {
		t.Fatalf("\nExpected %+v\nGot: %+v\n", expectedSet, set)
	}

	if !reflect.DeepEqual(expectedUnSet, unSet) {
		t.Fatalf("\nExpected %+v\nGot: %+v\n", expectedUnSet, unSet)
	}
}

func TestCheckLabel(t *testing.T) {

	darwin := map[string]bool{"darwin": true}
	linux := map[string]bool{"linux": true}
	both := map[string]bool{"darwin": true, "linux": true}

	darwinUser := RunConfig{Labels: darwin}
	linuxUser := RunConfig{Labels: linux}

	if !CheckLabel(nil, nil, RunConfig{}) {
		t.Fatalf("Test with no labels on host with no labels doesn't run")
	}

	if !CheckLabel(nil, nil, darwinUser) {
		t.Fatalf("Test with no labels doesn't run!")
	}

	if !CheckLabel(darwin, nil, darwinUser) {
		t.Fatalf("Darwin test doesn't run for darwin user")
	}

	if !CheckLabel(linux, nil, linuxUser) {
		t.Fatalf("Linux test doesn't run for linux user")
	}

	if CheckLabel(darwin, nil, linuxUser) {
		t.Fatalf("Darwin test runs for linuxUser")
	}

	if !CheckLabel(both, nil, darwinUser) {
		t.Fatalf("Test on darwin/linux doesn't run for darwin user")
	}
}
