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
	l := p.List()
	for _, t := range l {
		fmt.Println("Name: " + t.Name() + " Summary: " + t.Summary)
	}
}
