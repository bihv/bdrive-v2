package handler

import (
	"context"
	"encoding/json"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/biho/onedrive/internal/dto"
	"github.com/biho/onedrive/internal/service"
	"github.com/biho/onedrive/pkg/storage"
	"github.com/biho/onedrive/pkg/validator"
)

// ItemHandler handles HTTP requests for items (files and folders).
type ItemHandler struct {
	itemService *service.ItemService
	b2Client    *storage.B2Client
	validator   *validator.Validator
	log         *zap.Logger
}

// NewItemHandler creates a new ItemHandler.
func NewItemHandler(
	itemService *service.ItemService,
	b2Client *storage.B2Client,
	validator *validator.Validator,
	log *zap.Logger,
) *ItemHandler {
	return &ItemHandler{
		itemService: itemService,
		b2Client:    b2Client,
		validator:   validator,
		log:         log,
	}
}

// CreateFolder handles POST /api/v1/items/folder
func (h *ItemHandler) CreateFolder(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	var req dto.CreateFolderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid request body", Code: "INVALID_BODY",
		})
	}

	if errs := h.validator.Validate(&req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false, "errors": errs, "code": "VALIDATION_ERROR",
		})
	}

	item, err := h.itemService.CreateFolder(userID, &req)
	if err != nil {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		switch err.Error() {
		case "a folder with this name already exists":
			status = fiber.StatusConflict
			code = "NAME_EXISTS"
		case "parent folder not found":
			status = fiber.StatusNotFound
			code = "PARENT_NOT_FOUND"
		case "parent is not a folder":
			status = fiber.StatusBadRequest
			code = "PARENT_NOT_FOLDER"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: code,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse{
		Success: true,
		Data:    service.ToItemResponse(item, 0),
	})
}

// ListItems handles GET /api/v1/items?parent_id=
func (h *ItemHandler) ListItems(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	var parentID *uuid.UUID
	if pid := c.Query("parent_id"); pid != "" {
		parsed, err := uuid.Parse(pid)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Success: false, Error: "Invalid parent_id", Code: "INVALID_PARAM",
			})
		}
		parentID = &parsed
	}

	items, err := h.itemService.ListItems(userID, parentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to list items", Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    service.ToItemResponseList(items),
	})
}

// ListRecentItems handles GET /api/v1/items/recent
func (h *ItemHandler) ListRecentItems(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	limit := c.QueryInt("limit", 20)
	items, err := h.itemService.ListRecentItems(userID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to list recent items", Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    items,
	})
}

// ListStarredItems handles GET /api/v1/items/starred
func (h *ItemHandler) ListStarredItems(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	items, err := h.itemService.ListStarredItems(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to list starred items", Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    items,
	})
}

// GetItem handles GET /api/v1/items/:id
func (h *ItemHandler) GetItem(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	item, childCount, err := h.itemService.GetItem(id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "item not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: "NOT_FOUND",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    service.ToItemResponse(item, childCount),
	})
}

// UpdateItem handles PUT /api/v1/items/:id
func (h *ItemHandler) UpdateItem(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	var req dto.UpdateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid request body", Code: "INVALID_BODY",
		})
	}

	if errs := h.validator.Validate(&req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false, "errors": errs, "code": "VALIDATION_ERROR",
		})
	}

	item, err := h.itemService.UpdateItem(id, userID, &req)
	if err != nil {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		switch err.Error() {
		case "item not found":
			status = fiber.StatusNotFound
			code = "NOT_FOUND"
		case "an item with this name already exists":
			status = fiber.StatusConflict
			code = "NAME_EXISTS"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: code,
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    service.ToItemResponse(item, 0),
	})
}

// AddStar handles POST /api/v1/items/:id/star
func (h *ItemHandler) AddStar(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	if err := h.itemService.AddStar(userID, id); err != nil {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		if err.Error() == "item not found" {
			status = fiber.StatusNotFound
			code = "NOT_FOUND"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: code,
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    fiber.Map{"message": "Item starred"},
	})
}

// RemoveStar handles DELETE /api/v1/items/:id/star
func (h *ItemHandler) RemoveStar(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	if err := h.itemService.RemoveStar(userID, id); err != nil {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		if err.Error() == "item not found" {
			status = fiber.StatusNotFound
			code = "NOT_FOUND"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: code,
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    fiber.Map{"message": "Item unstarred"},
	})
}

// TrackItemActivity handles POST /api/v1/items/:id/activity
func (h *ItemHandler) TrackItemActivity(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	var req dto.TrackItemActivityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid request body", Code: "INVALID_BODY",
		})
	}

	if errs := h.validator.Validate(&req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false, "errors": errs, "code": "VALIDATION_ERROR",
		})
	}

	if err := h.itemService.TrackItemActivity(userID, id, req.Type); err != nil {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		switch err.Error() {
		case "item not found":
			status = fiber.StatusNotFound
			code = "NOT_FOUND"
		case "recent is only supported for files":
			status = fiber.StatusBadRequest
			code = "IS_FOLDER"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: code,
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    fiber.Map{"message": "Activity tracked"},
	})
}

// UpdateItemContent handles PUT /api/v1/items/:id/content
// Overwrites the content of an existing file with the uploaded data
func (h *ItemHandler) UpdateItemContent(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	// Get file from form data
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "No file uploaded", Code: "NO_FILE",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to open file", Code: "FILE_ERROR",
		})
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to read file", Code: "FILE_READ_ERROR",
		})
	}

	err = h.itemService.UpdateFileContent(context.Background(), userID, id.String(), content)
	if err != nil {
		status := fiber.StatusInternalServerError
		code := "UPDATE_ERROR"
		switch err.Error() {
		case "item not found":
			status = fiber.StatusNotFound
			code = "NOT_FOUND"
		case "item is not a valid file":
			status = fiber.StatusBadRequest
			code = "IS_FOLDER"
		case "storage not configured":
			status = fiber.StatusServiceUnavailable
			code = "STORAGE_NOT_CONFIGURED"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: code,
		})
	}

	// Fetch updated item to return
	item, _, err := h.itemService.GetItem(id, userID)
	if err != nil {
		// Even if we fail to fetch, the update succeeded.
		// Return a plain success.
		return c.JSON(dto.SuccessResponse{
			Success: true,
			Data:    fiber.Map{"message": "File content updated successfully"},
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    service.ToItemResponse(item, 0),
	})
}

// DeleteItem handles DELETE /api/v1/items/:id
func (h *ItemHandler) DeleteItem(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	if err := h.itemService.DeleteItem(id, userID); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "item not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: "NOT_FOUND",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    fiber.Map{"message": "Item deleted successfully"},
	})
}

// GetFolderTree handles GET /api/v1/items/tree
func (h *ItemHandler) GetFolderTree(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	tree, err := h.itemService.GetFolderTree(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to get folder tree", Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    tree,
	})
}

// Search handles GET /api/v1/items/search?q=<query>&limit=<limit>
// Returns items matching the query, sorted by relevance.
func (h *ItemHandler) Search(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Missing query parameter 'q'", Code: "INVALID_PARAM",
		})
	}

	limit := c.QueryInt("limit", 20)
	if limit <= 0 {
		limit = 20
	}

	items, err := h.itemService.SearchItems(userID, query, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to search items", Code: "INTERNAL_ERROR",
		})
	}

	// Convert to search result response
	results := make([]*dto.SearchResultResponse, 0, len(items))
	for i := range items {
		results = append(results, service.ToSearchResultResponse(&items[i]))
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    results,
	})
}

// UploadFile handles POST /api/v1/items/upload
// Supports multipart form data upload
func (h *ItemHandler) UploadFile(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	// Check if B2 client is configured
	if h.b2Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(dto.ErrorResponse{
			Success: false, Error: "Storage not configured", Code: "STORAGE_NOT_CONFIGURED",
		})
	}

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "No file provided", Code: "NO_FILE",
		})
	}

	// Get filename from form or use the uploaded filename
	filename := c.FormValue("name", file.Filename)
	if filename == "" {
		filename = file.Filename
	}

	// Get parent_id if provided
	var parentID *uuid.UUID
	if pid := c.FormValue("parent_id"); pid != "" {
		parsed, err := uuid.Parse(pid)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Success: false, Error: "Invalid parent_id", Code: "INVALID_PARAM",
			})
		}
		parentID = &parsed
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to read file", Code: "FILE_READ_ERROR",
		})
	}
	defer src.Close()

	// Get file size
	fileSize := file.Size

	// Call service to upload file
	item, err := h.itemService.UploadFile(context.Background(), userID, parentID, filename, fileSize, src)
	if err != nil {
		status := fiber.StatusInternalServerError
		code := "UPLOAD_ERROR"
		switch err.Error() {
		case "parent folder not found":
			status = fiber.StatusNotFound
			code = "PARENT_NOT_FOUND"
		case "parent is not a folder":
			status = fiber.StatusBadRequest
			code = "PARENT_NOT_FOLDER"
		case "a file with this name already exists":
			status = fiber.StatusConflict
			code = "NAME_EXISTS"
		case "storage not configured":
			status = fiber.StatusServiceUnavailable
			code = "STORAGE_NOT_CONFIGURED"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: code,
		})
	}

	// Generate download URL (optional, expires in 1 hour)
	var downloadURL *string
	if item.StorageKey != nil {
		url, err := h.b2Client.GetFileURL(context.Background(), *item.StorageKey, 3600)
		if err == nil {
			downloadURL = &url
		}
	}

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse{
		Success: true,
		Data:    service.ToUploadResponse(item, downloadURL),
	})
}

// GetPreSignedUploadURL handles POST /api/v1/items/upload-url
// Returns a pre-signed URL for direct upload to B2
func (h *ItemHandler) GetPreSignedUploadURL(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	// Check if B2 client is configured
	if h.b2Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(dto.ErrorResponse{
			Success: false, Error: "Storage not configured", Code: "STORAGE_NOT_CONFIGURED",
		})
	}

	var req dto.GetUploadURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid request body", Code: "INVALID_BODY",
		})
	}

	if errs := h.validator.Validate(&req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false, "errors": errs, "code": "VALIDATION_ERROR",
		})
	}

	result, err := h.itemService.GetPreSignedUploadURL(context.Background(), userID, &req)
	if err != nil {
		status := fiber.StatusInternalServerError
		code := "UPLOAD_ERROR"
		switch err.Error() {
		case "parent folder not found":
			status = fiber.StatusNotFound
			code = "PARENT_NOT_FOUND"
		case "parent is not a folder":
			status = fiber.StatusBadRequest
			code = "PARENT_NOT_FOLDER"
		case "a file with this name already exists":
			status = fiber.StatusConflict
			code = "NAME_EXISTS"
		case "storage not configured":
			status = fiber.StatusServiceUnavailable
			code = "STORAGE_NOT_CONFIGURED"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: code,
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    result,
	})
}

// GetUploadPartURL handles POST /api/v1/items/upload-part-url
// Returns a pre-signed URL for uploading a part in multipart upload
func (h *ItemHandler) GetUploadPartURL(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	if h.b2Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(dto.ErrorResponse{
			Success: false, Error: "Storage not configured", Code: "STORAGE_NOT_CONFIGURED",
		})
	}

	var req struct {
		StorageKey string `json:"storage_key" validate:"required"`
		UploadID   string `json:"upload_id" validate:"required"`
		PartNumber int    `json:"part_number" validate:"required,min=1"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid request body", Code: "INVALID_BODY",
		})
	}

	uploadURL, err := h.itemService.GetUploadPartURL(context.Background(), userID, req.StorageKey, req.UploadID, req.PartNumber)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: "UPLOAD_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data: fiber.Map{
			"upload_url": uploadURL,
		},
	})
}

// CompleteLargeUpload handles POST /api/v1/items/complete-upload
// Completes a multipart upload
func (h *ItemHandler) CompleteLargeUpload(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	if h.b2Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(dto.ErrorResponse{
			Success: false, Error: "Storage not configured", Code: "STORAGE_NOT_CONFIGURED",
		})
	}

	var req dto.CompleteLargeUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid request body", Code: "INVALID_BODY",
		})
	}

	if errs := h.validator.Validate(&req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false, "errors": errs, "code": "VALIDATION_ERROR",
		})
	}

	result, err := h.itemService.CompleteLargeUpload(context.Background(), userID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: "UPLOAD_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    result,
	})
}

// ListTrash handles GET /api/v1/trash
func (h *ItemHandler) ListTrash(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	items, err := h.itemService.ListTrash(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to list trash", Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    items,
	})
}

// RestoreItem handles POST /api/v1/trash/:id/restore
func (h *ItemHandler) RestoreItem(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	var req dto.RestoreItemRequest
	if body := c.Body(); len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Success: false, Error: "Invalid request body", Code: "INVALID_PARAM",
			})
		}
	}

	item, err := h.itemService.RestoreItem(id, userID, &req)
	if err != nil {
		if trashErr, ok := err.(*service.TrashError); ok {
			if trashErr.Code == "PARENT_DELETED" || trashErr.Code == "NAME_CONFLICT" {
				return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
					Success: false, Error: trashErr.Message, Code: trashErr.Code,
				})
			}
		}
		status := fiber.StatusInternalServerError
		if err.Error() == "item not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    item,
	})
}

// PermanentDeleteItem handles DELETE /api/v1/trash/:id
func (h *ItemHandler) PermanentDeleteItem(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	if err := h.itemService.PermanentDeleteItem(id, userID); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "item not found" {
			status = fiber.StatusNotFound
		}
		if err.Error() == "storage service unavailable" {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    fiber.Map{"message": "Item permanently deleted"},
	})
}

// GetPreview handles GET /api/v1/items/:id/preview
// Returns a pre-signed URL for file preview.
func (h *ItemHandler) GetPreview(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	item, _, err := h.itemService.GetItem(id, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "item not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: "NOT_FOUND",
		})
	}

	if item.IsFolder {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Cannot preview a folder", Code: "IS_FOLDER",
		})
	}

	if item.StorageKey == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success: false, Error: "File has no storage key", Code: "NO_STORAGE",
		})
	}

	if h.b2Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(dto.ErrorResponse{
			Success: false, Error: "Storage not configured", Code: "STORAGE_NOT_CONFIGURED",
		})
	}

	// Generate pre-signed URL (1 hour expiry)
	url, err := h.b2Client.GetFileURL(context.Background(), *item.StorageKey, 3600)
	if err != nil {
		h.log.Error("Failed to generate preview URL", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to generate preview URL", Code: "URL_ERROR",
		})
	}

	mimeType := ""
	if item.MimeType != nil {
		mimeType = *item.MimeType
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data: fiber.Map{
			"url":        url,
			"mime_type":  mimeType,
			"name":       item.Name,
			"size":       item.Size,
			"updated_at": item.UpdatedAt,
		},
	})
}
