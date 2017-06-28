package local

import (
	"fmt"
	"testing"
)

func TestFindingTests(t *testing.T) {
	p, err := NewProject("testdata/cases")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Init(); err != nil {
		t.Fatal(err)
	}

	expected := []Result{
		{Name: "test.osx"},
		{Name: "test.win"},
		{Name: "test.apps.test"},
		{Name: "test.apps.basic.test"},
		{Name: "test.apps.advanced.test"},
	}

	config := RunConfig{}
	l := p.List(config)
	for i, tst := range l {
		if expected[i].Name != tst.Name {
			t.Fatalf("Error in test ordering:\n Got %+v\nExpected %+v\n", tst, expected[i])
		}
	}
}

func TestTestPattern(t *testing.T) {
	p, err := NewProject("testdata/cases")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Init(); err != nil {
		t.Fatal(err)
	}
	config := RunConfig{TestPattern: "test.apps.basic"}
	l := p.List(config)
	for _, tst := range l {
		if tst.Name == "test.apps.advanced.test" && tst.TestResult != Skip {
			t.Fatal("test.apps.advanced.test does not match the TestPattern test.apps.basic")
		}
		if tst.Name == "test.apps.basic.test" && tst.TestResult != Pass {
			t.Fatal("test.apps.basic.test matches the TestPattern but is not going to run")
		}
		fmt.Printf("Name: %s Summary: %s WillRun: %d\n", tst.Name, tst.Summary, tst.TestResult)
	}
}
