package persistence

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"api/internal/domain"
)

func TestUniqueTagIDsRemovesDuplicatesAndKeepsOrder(t *testing.T) {
	got := uniqueTagIDs([]uint{3, 1, 3, 2, 1})
	want := []uint{3, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("uniqueTagIDs = %v, want %v", got, want)
	}
	if got := uniqueTagIDs(nil); len(got) != 0 {
		t.Fatalf("uniqueTagIDs(nil) = %v, want empty", got)
	}
}

func TestPostRepositoryMySQLRejectsUnknownTagIDsAndRollsBack(t *testing.T) {
	db, userID, categoryID := newPostSlugMySQLTestDB(t)
	repository := &PostRepository{db: db}

	tag := TagModel{Name: "Real", Slug: fmt.Sprintf("real-%d", time.Now().UnixNano())}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}

	_, err := repository.Create(context.Background(), newTestPost(t, userID, categoryID, "post-with-bad-tag"), []uint{tag.ID, 999999})
	if !errors.Is(err, domain.ErrPostTagsNotFound) {
		t.Fatalf("create with unknown tag error = %v, want ErrPostTagsNotFound", err)
	}
	var count int64
	if err := db.Model(&PostModel{}).Where("slug = ?", "post-with-bad-tag").Count(&count).Error; err != nil {
		t.Fatalf("count posts: %v", err)
	}
	if count != 0 {
		t.Fatalf("post was saved despite unknown tag id (count=%d), transaction should roll back", count)
	}

	// id trùng không được coi là thiếu tag.
	created, err := repository.Create(context.Background(), newTestPost(t, userID, categoryID, "post-with-dup-tag"), []uint{tag.ID, tag.ID})
	if err != nil {
		t.Fatalf("create with duplicated tag id: %v", err)
	}
	if got := len(created.Data().Tags); got != 1 {
		t.Fatalf("tags = %d, want 1", got)
	}
}
