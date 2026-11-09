package usecase

import (
	"api/internal/domain"
	"context"
	"sync"
	"testing"
	"time"
)

// dueRepository mô phỏng đúng hợp đồng của PublishIfDue: với mỗi bài chỉ MỘT lần gọi
// được trả về true (giống UPDATE có điều kiện trong MySQL), các lần sau trả về false.
type dueRepository struct {
	domain.PostRepository
	mu        sync.Mutex
	due       []domain.Post
	scheduled map[uint]bool
}

func (r *dueRepository) ListDue(context.Context, time.Time, int) ([]domain.Post, error) {
	return r.due, nil
}

func (r *dueRepository) PublishIfDue(_ context.Context, id uint, _ time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.scheduled[id] {
		return false, nil
	}
	r.scheduled[id] = false
	return true, nil
}

type recordingAuditRepository struct {
	domain.AuditRepository
	mu      sync.Mutex
	actions []string
}

func (r *recordingAuditRepository) RecordLog(_ context.Context, _ uint, action, _, _, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actions = append(r.actions, action)
	return nil
}

func newDueUseCase(scheduled bool) (*PostUseCase, *recordingAuditRepository) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	post := domain.ReconstitutePost(domain.PostData{
		ID: 1, Title: "Scheduled", Slug: "scheduled", AuthorID: 7,
		Status: domain.PostStatusScheduled, ScheduledAt: &at,
	})
	posts := &dueRepository{due: []domain.Post{*post}, scheduled: map[uint]bool{1: scheduled}}
	audits := &recordingAuditRepository{}
	useCase := NewPostUseCase(posts, postUpdateUnitOfWork{repositories: TransactionRepositories{
		Posts: posts, Audits: audits,
	}})
	return useCase, audits
}

func TestPublishDuePostsPublishesAndAuditsOnce(t *testing.T) {
	useCase, audits := newDueUseCase(true)

	count, err := useCase.PublishDuePosts(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("PublishDuePosts returned error: %v", err)
	}
	if count != 1 || len(audits.actions) != 1 || audits.actions[0] != "AUTO_PUBLISH_POST" {
		t.Fatalf("count=%d audits=%v, want 1 post and one AUTO_PUBLISH_POST audit", count, audits.actions)
	}
}

func TestPublishDuePostsConcurrentRunsPublishEachPostOnce(t *testing.T) {
	// Nhiều instance cùng thấy bài trong ListDue; chỉ một instance được đăng và ghi audit.
	useCase, audits := newDueUseCase(true)

	const runners = 8
	counts := make(chan int, runners)
	var wg sync.WaitGroup
	for i := 0; i < runners; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count, err := useCase.PublishDuePosts(context.Background(), time.Now())
			if err != nil {
				t.Errorf("PublishDuePosts returned error: %v", err)
			}
			counts <- count
		}()
	}
	wg.Wait()
	close(counts)

	total := 0
	for count := range counts {
		total += count
	}
	if total != 1 {
		t.Fatalf("total published across runners = %d, want 1", total)
	}
	if len(audits.actions) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audits.actions))
	}
}

func TestPublishDuePostsSkipsPostThatInstanceLost(t *testing.T) {
	// Bài đã bị instance khác đăng (hoặc vừa bị sửa) giữa lúc ListDue và UPDATE.
	useCase, audits := newDueUseCase(false)

	count, err := useCase.PublishDuePosts(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("PublishDuePosts returned error: %v", err)
	}
	if count != 0 || len(audits.actions) != 0 {
		t.Fatalf("count=%d audits=%v, want nothing published and no audit", count, audits.actions)
	}
}
