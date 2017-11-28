package local

import (
	"reflect"
	"testing"
)

func TestSetEnv(t *testing.T) {

	env := []string{"FOO=foo", "BAR=bar"}

	setEnv(&env, "FOO", "baz")
	exp := []string{"FOO=baz", "BAR=bar"}
	if !reflect.DeepEqual(env, exp) {
		t.Fatalf("Overwriting an existing variable failed: %v != %v", env, exp)
	}

	setEnv(&env, "BAZ", "baz")
	exp = []string{"FOO=baz", "BAR=bar", "BAZ=baz"}
	if !reflect.DeepEqual(env, exp) {
		t.Fatalf("Adding an existing variable failed: %v != %v", env, exp)
	}

	setEnv(&env, "BAZ", "foo")
	exp = []string{"FOO=baz", "BAR=bar", "BAZ=foo"}
	if !reflect.DeepEqual(env, exp) {
		t.Fatalf("Overwriting an added variable failed: %v != %v", env, exp)
	}

	env = []string{"FOO_foo", "BAR=bar"}
	setEnv(&env, "FOO", "foo")
	exp = []string{"FOO_foo", "BAR=bar", "FOO=foo"}
	if !reflect.DeepEqual(env, exp) {
		t.Fatalf("Adding a variable to a malformed environment failed: %v != %v", env, exp)
	}
}
