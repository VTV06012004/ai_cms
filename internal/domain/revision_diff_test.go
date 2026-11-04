package domain

import "testing"

func TestDiffLinesMarksAddedRemovedAndUnchanged(t *testing.T) {
	lines := DiffLines("a\nb\nc", "a\nx\nc\nd")
	want := []DiffLine{
		{DiffEqual, "a"}, {DiffRemove, "b"}, {DiffAdd, "x"}, {DiffEqual, "c"}, {DiffAdd, "d"},
	}
	if len(lines) != len(want) {
		t.Fatalf("got %+v, want %+v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("line %d = %+v, want %+v", i, lines[i], want[i])
		}
	}
}

func TestDiffLinesHandlesEmptyAndWindowsLineEndings(t *testing.T) {
	if got := DiffLines("", ""); len(got) != 0 {
		t.Fatalf("expected no lines, got %+v", got)
	}
	got := DiffLines("", "one\r\ntwo")
	if len(got) != 2 || got[0] != (DiffLine{DiffAdd, "one"}) || got[1] != (DiffLine{DiffAdd, "two"}) {
		t.Fatalf("unexpected diff: %+v", got)
	}
	if got := DiffLines("same\r\ntext", "same\ntext"); len(got) != 2 || got[0].Op != DiffEqual || got[1].Op != DiffEqual {
		t.Fatalf("CRLF must equal LF: %+v", got)
	}
}

func TestCompareRevisionsReportsFieldChangesAndCounts(t *testing.T) {
	from := RevisionData{Title: "Old", Slug: "s", Content: "a\nb", Status: PostStatusDraft, CategoryID: 1}
	to := RevisionData{Title: "New", Slug: "s", Content: "a\nc\nd", Status: PostStatusPublished, CategoryID: 2}

	result := CompareRevisions(from, to)

	fields := map[string]FieldChange{}
	for _, change := range result.Changes {
		fields[change.Field] = change
	}
	if len(fields) != 3 || fields["title"].To != "New" || fields["status"].From != "draft" ||
		fields["category_id"].From != "1" || fields["category_id"].To != "2" {
		t.Fatalf("unexpected changes: %+v", result.Changes)
	}
	if _, changed := fields["slug"]; changed {
		t.Fatal("unchanged slug must not be reported")
	}
	if result.Added != 2 || result.Removed != 1 {
		t.Fatalf("added=%d removed=%d, want 2 and 1", result.Added, result.Removed)
	}
}
