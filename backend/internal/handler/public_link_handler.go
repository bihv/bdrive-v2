package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/biho/onedrive/internal/dto"
	"github.com/biho/onedrive/internal/service"
	"github.com/biho/onedrive/pkg/validator"
)

type PublicLinkHandler struct {
	publicLinkService *service.PublicLinkService
	validator         *validator.Validator
	log               *zap.Logger
}

func NewPublicLinkHandler(publicLinkService *service.PublicLinkService, validator *validator.Validator, log *zap.Logger) *PublicLinkHandler {
	return &PublicLinkHandler{
		publicLinkService: publicLinkService,
		validator:         validator,
		log:               log,
	}
}

func (h *PublicLinkHandler) ListItemPublicLinks(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	itemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	links, err := h.publicLinkService.ListItemPublicLinks(userID, itemID)
	if err != nil {
		return h.writePublicLinkError(c, err)
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    links,
	})
}

func (h *PublicLinkHandler) CreatePublicLink(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	itemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	var req dto.CreatePublicLinkRequest
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

	link, err := h.publicLinkService.CreatePublicLink(userID, itemID, &req)
	if err != nil {
		return h.writePublicLinkError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse{
		Success: true,
		Data:    link,
	})
}

func (h *PublicLinkHandler) UpdatePublicLink(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	publicLinkID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid public link ID", Code: "INVALID_PARAM",
		})
	}

	var req dto.UpdatePublicLinkRequest
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

	link, err := h.publicLinkService.UpdatePublicLink(userID, publicLinkID, &req, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return h.writePublicLinkError(c, err)
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    link,
	})
}

func (h *PublicLinkHandler) RevokePublicLink(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	publicLinkID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid public link ID", Code: "INVALID_PARAM",
		})
	}

	link, err := h.publicLinkService.RevokePublicLink(userID, publicLinkID, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return h.writePublicLinkError(c, err)
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    link,
	})
}

func (h *PublicLinkHandler) GetPublicLink(c *fiber.Ctx) error {
	detail, err := h.publicLinkService.GetPublicLinkDetail(c.Params("token"), extractPublicLinkSession(c), c.IP(), c.Get("User-Agent"))
	if err != nil {
		return h.writePublicLinkError(c, err)
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    detail,
	})
}

func (h *PublicLinkHandler) AuthenticatePublicLink(c *fiber.Ctx) error {
	var req dto.AuthenticatePublicLinkRequest
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

	response, err := h.publicLinkService.Authenticate(c.Params("token"), req.Password, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return h.writePublicLinkError(c, err)
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

func (h *PublicLinkHandler) ListSharedItems(c *fiber.Ctx) error {
	var parentID *uuid.UUID
	if rawParentID := strings.TrimSpace(c.Query("parent_id")); rawParentID != "" {
		parsed, err := uuid.Parse(rawParentID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Success: false, Error: "Invalid parent_id", Code: "INVALID_PARAM",
			})
		}
		parentID = &parsed
	}

	response, err := h.publicLinkService.ListSharedItems(c.Params("token"), extractPublicLinkSession(c), parentID, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return h.writePublicLinkError(c, err)
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

func (h *PublicLinkHandler) StreamSharedItem(c *fiber.Ctx) error {
	var itemID *uuid.UUID
	if rawItemID := strings.TrimSpace(c.Query("item_id")); rawItemID != "" {
		parsed, err := uuid.Parse(rawItemID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Success: false, Error: "Invalid item_id", Code: "INVALID_PARAM",
			})
		}
		itemID = &parsed
	}

	result, err := h.publicLinkService.StreamSharedItem(
		c.Context(),
		c.Params("token"),
		extractPublicLinkSession(c),
		itemID,
		c.QueryBool("download", false),
		c.Get("Range"),
		c.IP(),
		c.Get("User-Agent"),
	)
	if err != nil {
		return h.writePublicLinkError(c, err)
	}

	// Public preview needs to be embeddable from the frontend origin.
	c.Response().Header.Del("X-Frame-Options")
	c.Set("Content-Type", result.ContentType)
	c.Set("Content-Disposition", h.publicLinkService.BuildContentDisposition(result.FileName, c.QueryBool("download", false)))
	c.Set("Cache-Control", "no-store")
	c.Set("Accept-Ranges", "bytes")

	if result.Stream.ETag != nil {
		c.Set("ETag", *result.Stream.ETag)
	}
	if result.Stream.LastModified != nil {
		c.Set("Last-Modified", result.Stream.LastModified.UTC().Format(http.TimeFormat))
	}
	if result.Stream.ContentRange != nil {
		c.Set("Content-Range", *result.Stream.ContentRange)
		c.Status(fiber.StatusPartialContent)
	}
	if result.Stream.ContentLength >= 0 {
		c.Set("Content-Length", strconv.FormatInt(result.Stream.ContentLength, 10))
		return c.SendStream(result.Stream.Body, int(result.Stream.ContentLength))
	}

	return c.SendStream(result.Stream.Body)
}

func (h *PublicLinkHandler) writePublicLinkError(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := err.Error()

	switch {
	case strings.Contains(err.Error(), "invalid expires_at"):
		status = fiber.StatusBadRequest
		code = "INVALID_EXPIRES_AT"
	case strings.Contains(err.Error(), "expires_at must be in the future"):
		status = fiber.StatusBadRequest
		code = "INVALID_EXPIRES_AT"
	case strings.Contains(err.Error(), "password is required when enabling protection"):
		status = fiber.StatusBadRequest
		code = "PASSWORD_REQUIRED"
	case err == service.ErrSharedItemNotFound:
		status = fiber.StatusNotFound
		code = "NOT_FOUND"
	case err == service.ErrPublicLinkNotFound:
		status = fiber.StatusNotFound
		code = "PUBLIC_LINK_NOT_FOUND"
	case err == service.ErrPublicLinkExpired:
		status = fiber.StatusGone
		code = "PUBLIC_LINK_EXPIRED"
		message = "Public link has expired"
	case err == service.ErrPublicLinkRevoked:
		status = fiber.StatusGone
		code = "PUBLIC_LINK_REVOKED"
		message = "Public link has been revoked"
	case err == service.ErrPublicLinkAccessDenied:
		status = fiber.StatusForbidden
		code = "PUBLIC_LINK_ACCESS_DENIED"
		message = "Access denied"
	case err == service.ErrPublicLinkWrongPassword:
		status = fiber.StatusForbidden
		code = "WRONG_PASSWORD"
		message = "Wrong password"
	case err == service.ErrSharedItemNotFolder:
		status = fiber.StatusBadRequest
		code = "IS_NOT_FOLDER"
	case err == service.ErrSharedItemNotFile:
		status = fiber.StatusBadRequest
		code = "IS_FOLDER"
	case err == service.ErrStorageNotConfigured:
		status = fiber.StatusServiceUnavailable
		code = "STORAGE_NOT_CONFIGURED"
	case err == service.ErrFileHasNoStorageKey:
		status = fiber.StatusNotFound
		code = "NO_STORAGE"
	}

	if status == fiber.StatusInternalServerError {
		h.log.Error("Public link handler error", zap.Error(err))
	}

	return c.Status(status).JSON(dto.ErrorResponse{
		Success: false,
		Error:   message,
		Code:    code,
	})
}

func extractPublicLinkSession(c *fiber.Ctx) string {
	if token := strings.TrimSpace(c.Get("X-Public-Link-Session")); token != "" {
		return token
	}
	return strings.TrimSpace(c.Query("session"))
}
