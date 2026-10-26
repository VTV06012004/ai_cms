package domain

import (
	"strings"
	"time"
)

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusReview    PostStatus = "review"
	PostStatusApproved  PostStatus = "approved"
	PostStatusRejected  PostStatus = "rejected"
	PostStatusPublished PostStatus = "published"
)

type Post struct {
	id              uint
	title           string
	slug            string
	content         string
	excerpt         string
	status          PostStatus
	rejectionReason string
	authorID        uint
	author          *User
	moderatorID     *uint
	moderator       *User
	categoryID      uint
	category        *Category
	seoTitle        string
	seoDescription  string
	publishedAt     *time.Time
	createdAt       time.Time
	updatedAt       time.Time
	tags            []Tag
}

type PostData struct {
	ID              uint
	Title           string
	Slug            string
	Content         string
	Excerpt         string
	Status          PostStatus
	RejectionReason string
	AuthorID        uint
	Author          *User
	ModeratorID     *uint
	Moderator       *User
	CategoryID      uint
	Category        *Category
	SEOTitle        string
	SEODescription  string
	PublishedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Tags            []Tag
}

func NewPost(title, slug, content string, status PostStatus, categoryID, authorID uint) (*Post, error) {
	title = strings.TrimSpace(title)
	slug = strings.TrimSpace(slug)
	if title == "" {
		return nil, ErrPostTitleRequired
	}
	if slug == "" {
		return nil, ErrPostSlugRequired
	}
	if categoryID == 0 {
		return nil, ErrPostCategoryRequired
	}
	if status == "" {
		status = PostStatusDraft
	}
	if !isValidPostStatus(status) {
		return nil, ErrInvalidPostStatus
	}
	return &Post{
		title: title, slug: slug, content: content,
		status: status, categoryID: categoryID, authorID: authorID,
	}, nil
}

func ReconstitutePost(data PostData) *Post {
	var moderatorID *uint
	if data.ModeratorID != nil {
		id := *data.ModeratorID
		moderatorID = &id
	}
	var publishedAt *time.Time
	if data.PublishedAt != nil {
		at := *data.PublishedAt
		publishedAt = &at
	}
	var category *Category
	if data.Category != nil {
		categoryData := data.Category.Data()
		category = ReconstituteCategory(categoryData)
	}
	return &Post{
		id: data.ID, title: data.Title, slug: data.Slug, content: data.Content,
		excerpt: data.Excerpt, status: data.Status, rejectionReason: data.RejectionReason,
		authorID: data.AuthorID, author: cloneUser(data.Author), moderatorID: moderatorID,
		moderator: cloneUser(data.Moderator), categoryID: data.CategoryID, category: category,
		seoTitle: data.SEOTitle, seoDescription: data.SEODescription,
		publishedAt: publishedAt, createdAt: data.CreatedAt,
		updatedAt: data.UpdatedAt, tags: append([]Tag(nil), data.Tags...),
	}
}

func (p *Post) Data() PostData {
	var moderatorID *uint
	if p.moderatorID != nil {
		id := *p.moderatorID
		moderatorID = &id
	}
	var publishedAt *time.Time
	if p.publishedAt != nil {
		at := *p.publishedAt
		publishedAt = &at
	}
	var category *Category
	if p.category != nil {
		category = ReconstituteCategory(p.category.Data())
	}
	return PostData{
		ID: p.id, Title: p.title, Slug: p.slug, Content: p.content,
		Excerpt: p.excerpt, Status: p.status, RejectionReason: p.rejectionReason,
		AuthorID: p.authorID, Author: cloneUser(p.author), ModeratorID: moderatorID,
		Moderator: cloneUser(p.moderator), CategoryID: p.categoryID, Category: category,
		SEOTitle: p.seoTitle, SEODescription: p.seoDescription,
		PublishedAt: publishedAt, CreatedAt: p.createdAt,
		UpdatedAt: p.updatedAt, Tags: append([]Tag(nil), p.tags...),
	}
}

func (p *Post) CanBeManagedBy(userID uint, role UserRole) bool {
	switch role {
	case UserRoleAdmin, UserRoleEditor:
		return true
	case UserRoleAuthor:
		return p.authorID == userID
	default:
		return false
	}
}

func (p *Post) Update(title, slug, content string, status PostStatus, categoryID uint) error {
	title = strings.TrimSpace(title)
	slug = strings.TrimSpace(slug)
	if title == "" {
		return ErrPostTitleRequired
	}
	if slug == "" {
		return ErrPostSlugRequired
	}
	if categoryID == 0 {
		return ErrPostCategoryRequired
	}
	if status == "" {
		status = PostStatusDraft
	}
	if !isValidPostStatus(status) {
		return ErrInvalidPostStatus
	}
	p.title = title
	p.slug = slug
	p.content = content
	p.status = status
	p.categoryID = categoryID
	return nil
}

func (p *Post) Moderate(status PostStatus, reason string, moderatorID uint, at time.Time) error {
	if p.status != PostStatusReview {
		return ErrPostNotInReview
	}
	switch status {
	case PostStatusPublished:
		p.status = status
		p.rejectionReason = ""
		p.moderatorID = &moderatorID
		p.publishedAt = &at
	case PostStatusRejected:
		if reason == "" {
			return ErrRejectionReasonRequired
		}
		p.status = status
		p.rejectionReason = reason
		p.moderatorID = &moderatorID
	default:
		return ErrInvalidPostStatus
	}
	return nil
}

func isValidPostStatus(status PostStatus) bool {
	switch status {
	case PostStatusDraft, PostStatusReview, PostStatusApproved, PostStatusRejected, PostStatusPublished:
		return true
	default:
		return false
	}
}
