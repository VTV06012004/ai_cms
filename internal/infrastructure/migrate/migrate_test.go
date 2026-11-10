package migrate

import (
	"testing"
	"testing/fstest"
)

func TestLoadSortsByVersionAndIgnoresOtherFiles(t *testing.T) {
	migrations, err := Load(fstest.MapFS{
		"0002_add_x.sql": {Data: []byte("SELECT 2;")},
		"0001_init.sql":  {Data: []byte("SELECT 1;")},
		"embed.go":       {Data: []byte("package migrations")},
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(migrations) != 2 || migrations[0].Version != 1 || migrations[1].Version != 2 {
		t.Fatalf("unexpected order: %+v", migrations)
	}
	if migrations[0].Name != "0001_init" || migrations[0].SQL != "SELECT 1;" {
		t.Fatalf("unexpected first migration: %+v", migrations[0])
	}
}

func TestLoadRejectsBadNamesGapsAndDuplicates(t *testing.T) {
	cases := map[string]fstest.MapFS{
		"bad name":  {"init.sql": {Data: []byte("x")}},
		"gap":       {"0001_a.sql": {Data: []byte("x")}, "0003_c.sql": {Data: []byte("x")}},
		"not first": {"0002_b.sql": {Data: []byte("x")}},
		"duplicate": {"0001_a.sql": {Data: []byte("x")}, "0001_b.sql": {Data: []byte("y")}},
	}
	for name, fsys := range cases {
		if _, err := Load(fsys); err == nil {
			t.Errorf("%s: Load succeeded, expected an error", name)
		}
	}
}

func TestChecksumChangesWhenFileChanges(t *testing.T) {
	a, _ := Load(fstest.MapFS{"0001_init.sql": {Data: []byte("SELECT 1;")}})
	b, _ := Load(fstest.MapFS{"0001_init.sql": {Data: []byte("SELECT 1; ")}})
	if a[0].Checksum == b[0].Checksum {
		t.Fatal("checksum did not change when the file content changed")
	}
}

func TestVerifyAppliedDetectsEditedAndUnknownMigrations(t *testing.T) {
	migrations := []Migration{{Version: 1, Name: "0001_init", Checksum: "aaa"}}

	if err := verifyApplied(map[int]appliedRow{1: {Name: "0001_init", Checksum: "aaa"}}, migrations); err != nil {
		t.Fatalf("unchanged migration rejected: %v", err)
	}
	if err := verifyApplied(map[int]appliedRow{1: {Name: "0001_init", Checksum: "bbb"}}, migrations); err == nil {
		t.Fatal("edited migration was not detected")
	}
	if err := verifyApplied(map[int]appliedRow{2: {Name: "0002_new", Checksum: "ccc"}}, migrations); err == nil {
		t.Fatal("migration unknown to this binary was not detected")
	}
}
