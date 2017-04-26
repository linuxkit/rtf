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
