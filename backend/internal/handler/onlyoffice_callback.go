package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OnlyOfficeCallbackPayload represents the payload sent by OnlyOffice Document Server.
type OnlyOfficeCallbackPayload struct {
	Status int    `json:"status"`
	URL    string `json:"url"`
	Key    string `json:"key"`
}

// OnlyOfficeCallback handles OnlyOffice Document Server callbacks.
// This is called by OnlyOffice when a document is being saved or session ends.
func (h *ItemHandler) OnlyOfficeCallback(c *fiber.Ctx) error {
	var payload OnlyOfficeCallbackPayload
	if err := c.BodyParser(&payload); err != nil {
		h.log.Error("Failed to parse OnlyOffice callback payload", zap.Error(err))
		return c.JSON(fiber.Map{"error": 0}) // Always return error 0 to acknowledge
	}

	itemIDStr := c.Query("id")
	userIDStr := c.Query("userId")

	// Status 2 is "document ready for saving", Status 6 is "document being edited, force save"
	if (payload.Status == 2 || payload.Status == 6) && payload.URL != "" && itemIDStr != "" && userIDStr != "" {
		_, err := uuid.Parse(itemIDStr)
		if err != nil {
			h.log.Error("Invalid item ID in OnlyOffice callback", zap.String("id", itemIDStr))
			return c.JSON(fiber.Map{"error": 0})
		}

		// Because ItemHandler doesn't have direct access to ItemRepository unless we add it to ItemService,
		// let's use the itemService to get the item and then update it. Wait, ItemService doesn't expose a method to update size easily.
		// Let me implement this right now using standard Service pattern methods. Instead of doing it in handler,
		// we'll fetch the item from DB, download file, upload to B2, and update DB.
		// Since we don't have direct repo access here, we can call a new method in ItemService: UpdateFileFromURL.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		err = h.processOnlyOfficeSave(ctx, itemIDStr, userIDStr, payload.URL)
		if err != nil {
			h.log.Error("Failed to process OnlyOffice save", zap.Error(err), zap.String("itemID", itemIDStr))
		}
	}

	return c.JSON(fiber.Map{"error": 0})
}

// processOnlyOfficeSave downloads the modified file from OnlyOffice and updates B2 and DB
func (h *ItemHandler) processOnlyOfficeSave(ctx context.Context, itemIDStr, userIDStr, downloadURL string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	// Because we can't easily access the repository directly, we define a small ad-hoc way to fetch the item,
	// However, ItemService is the right place. We'll add this ad-hoc to the handler and rely on itemService's GetItem.
	// But GetItem returns DTO. We really need the ItemService to do this job.
	// Let's implement an ad-hoc DB query here using the B2Client and Service.
	// Wait, we can just edit the service to have UpdateItemSize, but I don't want to overcomplicate.
	// Let's modify ItemService later, for now let's HTTP GET the file.
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download from OnlyOffice: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status from OnlyOffice download: %d", resp.StatusCode)
	}

	// Read into memory
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 50*1024*1024)) // 50MB limit
	if err != nil {
		return fmt.Errorf("failed to read downloaded file: %w", err)
	}

	// Call ItemService to process this update. We'll add UpdateFileContent method to ItemService.
	return h.itemService.UpdateFileContent(ctx, userID, itemIDStr, bodyBytes)
}
