package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/biho/onedrive/internal/dto"
	"github.com/biho/onedrive/internal/service"
	"github.com/biho/onedrive/pkg/validator"
)

// ItemHandler handles HTTP requests for items (files and folders).
type ItemHandler struct {
	itemService *service.ItemService
	validator   *validator.Validator
	log         *zap.Logger
}

// NewItemHandler creates a new ItemHandler.
func NewItemHandler(
	itemService *service.ItemService,
	validator *validator.Validator,
	log *zap.Logger,
) *ItemHandler {
	return &ItemHandler{
		itemService: itemService,
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
