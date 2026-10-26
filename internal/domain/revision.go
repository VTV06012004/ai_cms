package domain

import "time"

type Revision struct {
	id         uint
	postID     uint
	title      string
	slug       string
	content    string
	status     PostStatus
	categoryID uint
	userID     uint
	user       *User
	createdAt  time.Time
}

type RevisionData struct {
	ID         uint
	PostID     uint
	Title      string
	Slug       string
	Content    string
	Status     PostStatus
	CategoryID uint
	UserID     uint
	User       *User
	CreatedAt  time.Time
}

func NewRevision(post PostData, userID uint) *Revision {
	return ReconstituteRevision(RevisionData{
		PostID: post.ID, Title: post.Title, Slug: post.Slug,
		Content: post.Content, Status: post.Status,
		CategoryID: post.CategoryID, UserID: userID,
	})
}

func ReconstituteRevision(data RevisionData) *Revision {
	return &Revision{
		id: data.ID, postID: data.PostID, title: data.Title, slug: data.Slug,
		content: data.Content, status: data.Status, categoryID: data.CategoryID,
		userID: data.UserID, user: cloneUser(data.User), createdAt: data.CreatedAt,
	}
}

func (r *Revision) Data() RevisionData {
	return RevisionData{
		ID: r.id, PostID: r.postID, Title: r.title, Slug: r.slug,
		Content: r.content, Status: r.status, CategoryID: r.categoryID,
		UserID: r.userID, User: cloneUser(r.user), CreatedAt: r.createdAt,
	}
}
