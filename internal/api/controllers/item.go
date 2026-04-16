package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"go-fx-template/internal/api/middlewares"
	"go-fx-template/internal/api/models"
	"go-fx-template/internal/config"
	"go-fx-template/internal/models"
	"go-fx-template/internal/utils/noop"
)

type ItemService interface {
	CreateItem(ctx context.Context, id uuid.UUID, req apiModels.CreateItemRequest) (models.Item, error)
	GetAllItemsForUser(ctx context.Context, id uuid.UUID, p apiModels.PaginationParams) ([]models.Item, int64, error)
	GetItemByID(ctx context.Context, id uuid.UUID, r uuid.UUID) (models.Item, error)
	UpdateItemByID(ctx context.Context, id uuid.UUID, r uuid.UUID, req apiModels.UpdateItemRequest) (models.Item, error)
	RemoveItemImage(ctx context.Context, itemID uuid.UUID, requestedByUserID uuid.UUID) error
	DeleteItemByID(ctx context.Context, id uuid.UUID, r uuid.UUID) error
}

type ImageService interface {
	Upload(ctx context.Context, entity string, entityID uuid.UUID, reader io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, storedValue string) error
}

type ItemController struct {
	router       *gin.Engine
	mw           *middlewares.Middlewares
	itemService  ItemService
	imageService ImageService
}

func NewItemController(e *gin.Engine, mw *middlewares.Middlewares, itmSvc ItemService, imgSvc ImageService) *ItemController {
	return &ItemController{router: e, mw: mw, itemService: itmSvc, imageService: imgSvc}
}

func (ctrl *ItemController) RegisterRoutes() {
	basePath := ctrl.router.Group(viper.GetString(config.ApiBasePath))
	itemRoutes := basePath.Group("/items")
	{
		authorizedItemRoutes := itemRoutes.Group("").Use(ctrl.mw.AuthMiddleware())
		{
			authorizedItemRoutes.POST("", ctrl.CreateItem)
			authorizedItemRoutes.GET("", ctrl.GetAllItemsForUser)
			authorizedItemRoutes.GET("/:item_id", ctrl.GetItem)
			authorizedItemRoutes.PATCH("/:item_id", ctrl.UpdateItem)
			authorizedItemRoutes.DELETE("/:item_id", ctrl.DeleteItem)
			{
				authorizedItemRoutes.PUT("/:item_id/image", ctrl.UpdateItemImage)
				authorizedItemRoutes.DELETE("/:item_id/image", ctrl.RemoveItemImage)
			}
		}
	}
}

// CreateItem
// @Summary Create item
// @Description Creates a new item for the authenticated user.
// @Tags Items
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param request body apiModels.CreateItemRequest true "Create item payload"
// @Success 201 {object} apiModels.ItemResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 500 {object} apiModels.APIError
// @Router /items [post]
func (ctrl *ItemController) CreateItem(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	var req apiModels.CreateItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	item, err := ctrl.itemService.CreateItem(ctx, userID, req)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, apiModels.ToItemResponse(item))
}

// GetAllItemsForUser
// @Summary List current user items
// @Description Returns a paginated list of items that belong to the authenticated user.
// @Tags Items
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} apiModels.PaginatedItemsResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 500 {object} apiModels.APIError
// @Router /items [get]
func (ctrl *ItemController) GetAllItemsForUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	var paginationParams apiModels.PaginationParams
	if err := ctx.ShouldBindQuery(&paginationParams); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	items, total, err := ctrl.itemService.GetAllItemsForUser(ctx, userID, paginationParams)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.NewPaginatedResponse(apiModels.ToItemsResponse(items), total, paginationParams))
}

// GetItem
// @Summary Get item by ID
// @Description Returns a single item owned by the authenticated user.
// @Tags Items
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param item_id path string true "Item ID"
// @Success 200 {object} apiModels.ItemResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /items/{item_id} [get]
func (ctrl *ItemController) GetItem(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	itemID, err := uuid.Parse(ctx.Param("item_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid item ID")
		return
	}

	item, err := ctrl.itemService.GetItemByID(ctx, itemID, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.ToItemResponse(item))
}

// UpdateItem
// @Summary Update item
// @Description Updates an item owned by the authenticated user.
// @Tags Items
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param item_id path string true "Item ID"
// @Param request body apiModels.UpdateItemRequest true "Update item payload"
// @Success 200 {object} apiModels.ItemResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /items/{item_id} [patch]
func (ctrl *ItemController) UpdateItem(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	itemID, err := uuid.Parse(ctx.Param("item_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid item ID")
		return
	}

	var req apiModels.UpdateItemRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	item, err := ctrl.itemService.UpdateItemByID(ctx, itemID, userID, req)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.ToItemResponse(item))
}

// DeleteItem
// @Summary Delete item
// @Description Deletes an item owned by the authenticated user.
// @Tags Items
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param item_id path string true "Item ID"
// @Success 204 "No Content"
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /items/{item_id} [delete]
func (ctrl *ItemController) DeleteItem(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	itemID, err := uuid.Parse(ctx.Param("item_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid item ID")
		return
	}

	if err = ctrl.itemService.DeleteItemByID(ctx, itemID, userID); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

// UpdateItemImage
// @Summary Update item image
// @Description Uploads and sets a new image for an item owned by the authenticated user.
// @Tags Items
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param item_id path string true "Item ID"
// @Param image formData file true "Item image"
// @Success 200 "OK"
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /items/{item_id}/image [put]
func (ctrl *ItemController) UpdateItemImage(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	itemID, err := uuid.Parse(ctx.Param("item_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid item ID")
		return
	}

	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "image file is required")
		return
	}

	// Note: if behind nginx, ensure client_max_body_size matches MinioMaxFileSize
	if fileHeader.Size > viper.GetInt64(config.MinioMaxFileSize)<<20 { // 10 << 20 = 10 * 1 048 576 = 10 485 760 bytes = 10 MB
		apiModels.Error(ctx, http.StatusBadRequest, fmt.Sprintf("image file too large (max %d MB)", viper.GetInt(config.MinioMaxFileSize)))
		return
	}

	switch fileHeader.Header.Get("Content-Type") {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		noop.Continue()
	default:
		apiModels.Error(ctx, http.StatusBadRequest, "unsupported image format (PNG, JPG, WEBP or GIF only)")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	item, err := ctrl.itemService.GetItemByID(ctx, itemID, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	if item.Image != nil {
		_ = ctrl.imageService.Delete(ctx, *item.Image)
	}

	imageURL, err := ctrl.imageService.Upload(ctx, "item", itemID, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	req := apiModels.UpdateItemRequest{Image: &imageURL}
	if _, err = ctrl.itemService.UpdateItemByID(ctx, itemID, userID, req); err != nil {
		_ = ctrl.imageService.Delete(ctx, imageURL)
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.Status(http.StatusOK)
}

// RemoveItemImage
// @Summary Remove item image
// @Description Removes the current image from an item owned by the authenticated user.
// @Tags Items
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param item_id path string true "Item ID"
// @Success 204 "No Content"
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /items/{item_id}/image [delete]
func (ctrl *ItemController) RemoveItemImage(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	itemID, err := uuid.Parse(ctx.Param("item_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid item ID")
		return
	}

	item, err := ctrl.itemService.GetItemByID(ctx, itemID, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	if item.Image == nil {
		ctx.Status(http.StatusNoContent)
		return
	}

	if err = ctrl.itemService.RemoveItemImage(ctx, itemID, userID); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	_ = ctrl.imageService.Delete(ctx, *item.Image)

	ctx.Status(http.StatusNoContent)
}
