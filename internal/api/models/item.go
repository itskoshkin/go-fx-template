package apiModels

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"go-fx-template/internal/models"
)

type CreateItemRequest struct {
	Title       string  `json:"title" binding:"required" example:"Book"`
	Description *string `json:"description" example:"A good book"`
}

func (req *CreateItemRequest) Validate() error {
	if strings.TrimSpace(req.Title) == "" {
		return errors.New("title is required")
	}
	if req.Description != nil && strings.TrimSpace(*req.Description) == "" {
		return errors.New("description must be not empty or omitted")
	}

	return nil
}

type UpdateItemRequest struct {
	Image       *string `json:"image" binding:"omitempty" example:"null"`
	Title       *string `json:"title" binding:"omitempty" example:"Red book"`
	Description *string `json:"description" example:"A good book with red cover"`
}

func (req *UpdateItemRequest) Validate() error {
	if req.Image != nil && strings.TrimSpace(*req.Image) == "" {
		return errors.New("image must be not empty or omitted")
	}
	if req.Title != nil && strings.TrimSpace(*req.Title) == "" {
		return errors.New("title must be not empty or omitted")
	}
	if req.Description != nil && strings.TrimSpace(*req.Description) == "" {
		return errors.New("description must be not empty or omitted")
	}

	return nil
}

type ItemResponse struct {
	ID          uuid.UUID `json:"id" example:"019d82d5-8b25-78aa-b269-2b840cca9a8c"`
	UserID      uuid.UUID `json:"user_id" example:"019d82cc-62f9-734f-8a91-9627afbb1e1d"`
	Image       *string   `json:"image" example:"http://localhost:9000/go-fx-template/items/019d82d5-8b25-78aa-b269-2b840cca9a8c/019d82d7-dd6c-7b52-9fd7-5e5be6d901b4"`
	Title       string    `json:"title" example:"Book"`
	Description *string   `json:"description" example:"A good book"`
	CreatedAt   time.Time `json:"created_at" example:"2026-04-04T00:00:00.000000+03:00"`
	UpdatedAt   time.Time `json:"updated_at" example:"2026-04-04T00:00:00.000000+03:00"`
}

type PaginatedItemsResponse struct {
	Total      int64          `json:"total_objects" example:"5"`
	Limit      int            `json:"limit_per_page" example:"10"`
	Page       int            `json:"current_page" example:"1"`
	TotalPages int            `json:"total_pages" example:"1"`
	Data       []ItemResponse `json:"data"`
}

func ToItemResponse(item models.Item) ItemResponse {
	return ItemResponse{
		ID:          item.ID,
		UserID:      item.UserID,
		Image:       item.Image,
		Title:       item.Title,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func ToItemsResponse(items []models.Item) []ItemResponse {
	result := make([]ItemResponse, len(items))
	for i, item := range items {
		result[i] = ToItemResponse(item)
	}
	return result
}
