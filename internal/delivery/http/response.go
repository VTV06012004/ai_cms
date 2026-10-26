package http

import (
	"time"

	"api/internal/domain"
)

type userResponse struct {
	ID         uint      `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	CanApprove bool      `json:"can_approve"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type categoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ParentID    *uint  `json:"parent_id"`
}

type tagResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type postResponse struct {
	ID              uint              `json:"id"`
	Title           string            `json:"title"`
	Slug            string            `json:"slug"`
	Content         string            `json:"content"`
	Excerpt         string            `json:"excerpt"`
	Status          string            `json:"status"`
	RejectionReason string            `json:"rejection_reason"`
	AuthorID        uint              `json:"author_id"`
	Author          *userResponse     `json:"author,omitempty"`
	ModeratorID     *uint             `json:"moderator_id"`
	Moderator       *userResponse     `json:"moderator,omitempty"`
	CategoryID      uint              `json:"category_id"`
	Category        *categoryResponse `json:"category,omitempty"`
	SEOTitle        string            `json:"seo_title"`
	SEODescription  string            `json:"seo_description"`
	PublishedAt     *time.Time        `json:"published_at"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Tags            []tagResponse     `json:"tags,omitempty"`
}

type menuItemResponse struct {
	ID        uint      `json:"id"`
	MenuID    uint      `json:"menu_id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	TargetID  uint      `json:"target_id"`
	URL       string    `json:"url"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type menuResponse struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	Location  string             `json:"location"`
	MenuItems []menuItemResponse `json:"menu_items,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type revisionResponse struct {
	ID         uint          `json:"id"`
	PostID     uint          `json:"post_id"`
	Title      string        `json:"title"`
	Slug       string        `json:"slug"`
	Content    string        `json:"content"`
	Status     string        `json:"status"`
	CategoryID uint          `json:"category_id"`
	UserID     uint          `json:"user_id"`
	User       *userResponse `json:"user,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

type auditLogResponse struct {
	ID        uint          `json:"id"`
	UserID    uint          `json:"user_id"`
	User      *userResponse `json:"user,omitempty"`
	Action    string        `json:"action"`
	Target    string        `json:"target"`
	Details   string        `json:"details"`
	IPAddress string        `json:"ip_address"`
	CreatedAt time.Time     `json:"created_at"`
}

func toUserResponse(user domain.User) userResponse {
	data := user.Data()
	return userResponse{
		ID: data.ID, Email: string(data.Email), Name: data.Name, Role: string(data.Role),
		CanApprove: data.CanApprove,
		CreatedAt:  data.CreatedAt, UpdatedAt: data.UpdatedAt,
	}
}

func toUserResponsePtr(user *domain.User) *userResponse {
	if user == nil {
		return nil
	}
	response := toUserResponse(*user)
	return &response
}

func toCategoryResponse(category domain.Category) categoryResponse {
	data := category.Data()
	return categoryResponse{
		ID: data.ID, Name: data.Name, Slug: data.Slug,
		Description: data.Description, ParentID: data.ParentID,
	}
}

func toTagResponse(tag domain.Tag) tagResponse {
	data := tag.Data()
	return tagResponse{
		ID: data.ID, Name: data.Name, Slug: data.Slug,
		CreatedAt: data.CreatedAt, UpdatedAt: data.UpdatedAt,
	}
}

func toPostResponse(post domain.Post) postResponse {
	data := post.Data()
	response := postResponse{
		ID: data.ID, Title: data.Title, Slug: data.Slug, Content: data.Content,
		Excerpt: data.Excerpt, Status: string(data.Status), RejectionReason: data.RejectionReason,
		AuthorID: data.AuthorID, Author: toUserResponsePtr(data.Author),
		ModeratorID: data.ModeratorID, Moderator: toUserResponsePtr(data.Moderator),
		CategoryID: data.CategoryID, SEOTitle: data.SEOTitle,
		SEODescription: data.SEODescription, PublishedAt: data.PublishedAt,
		CreatedAt: data.CreatedAt, UpdatedAt: data.UpdatedAt,
	}
	if data.Category != nil {
		category := toCategoryResponse(*data.Category)
		response.Category = &category
	}
	if len(data.Tags) > 0 {
		response.Tags = make([]tagResponse, 0, len(data.Tags))
		for _, tag := range data.Tags {
			response.Tags = append(response.Tags, toTagResponse(tag))
		}
	}
	return response
}

func toPostResponses(posts []domain.Post) []postResponse {
	responses := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, toPostResponse(post))
	}
	return responses
}

func toMenuResponse(menu domain.Menu) menuResponse {
	data := menu.Data()
	response := menuResponse{
		ID: data.ID, Name: data.Name, Location: data.Location,
		CreatedAt: data.CreatedAt, UpdatedAt: data.UpdatedAt,
	}
	if len(data.MenuItems) > 0 {
		response.MenuItems = make([]menuItemResponse, 0, len(data.MenuItems))
		for _, item := range data.MenuItems {
			itemData := item.Data()
			response.MenuItems = append(response.MenuItems, menuItemResponse{
				ID: itemData.ID, MenuID: itemData.MenuID, Title: itemData.Title, Type: itemData.Type,
				TargetID: itemData.TargetID, URL: itemData.URL, SortOrder: itemData.SortOrder,
				CreatedAt: itemData.CreatedAt, UpdatedAt: itemData.UpdatedAt,
			})
		}
	}
	return response
}

func toRevisionResponse(revision domain.Revision) revisionResponse {
	data := revision.Data()
	return revisionResponse{
		ID: data.ID, PostID: data.PostID, Title: data.Title,
		Slug: data.Slug, Content: data.Content, Status: string(data.Status),
		CategoryID: data.CategoryID, UserID: data.UserID,
		User: toUserResponsePtr(data.User), CreatedAt: data.CreatedAt,
	}
}

func toAuditLogResponse(log domain.AuditLog) auditLogResponse {
	data := log.Data()
	return auditLogResponse{
		ID: data.ID, UserID: data.UserID, User: toUserResponsePtr(data.User),
		Action: data.Action, Target: data.Target, Details: data.Details,
		IPAddress: data.IPAddress, CreatedAt: data.CreatedAt,
	}
}
