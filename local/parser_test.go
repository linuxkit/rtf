package local

import "testing"

func TestParseTags(t *testing.T) {
	eSummary := "A Test"
	eAuthor := "Dave Tucker <dt@docker.com> Rolf Neugebauer <rofl.neugebauer@docker.com>"
	eLabels := "foo, bar, !baz"
	eRepeat := 5

	tags, err := ParseTags("testdata/test.sh")
	if err != nil {
		t.Fatalf("Error parsing tags")
	}

	if eSummary != tags.Summary {
		t.Fatalf("\nExpected: %s \nGot: %s\n", eSummary, tags.Summary)
	}
	if eAuthor != tags.Author {
		t.Fatalf("\nExpected: %s \nGot: %s\n", eAuthor, tags.Author)
	}
	if eLabels != tags.Labels {
		t.Fatalf("\nExpected: %s \nGot: %s\n", eLabels, tags.Labels)
	}
	if eRepeat != tags.Repeat {
		t.Fatalf("\nExpected: %d \nGot: %d\n", eRepeat, tags.Repeat)
	}
}

func TestParseBadTags(t *testing.T) {
	_, err := ParseTags("testdata/bad_test.sh")
	if err == nil {
		t.Fatalf("Should have caused an error")
	}
	if err.Error() != "Field LABELS specified multiple times" {
		t.Fatalf("Wrong error message")
	}
}
