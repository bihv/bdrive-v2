package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/biho/onedrive/internal/dto"
	"github.com/biho/onedrive/internal/model"
	"github.com/biho/onedrive/internal/repository"
	hashpkg "github.com/biho/onedrive/pkg/hash"
	jwtpkg "github.com/biho/onedrive/pkg/jwt"
	"github.com/biho/onedrive/pkg/storage"
)

const (
	PublicLinkStatusActive  = "active"
	PublicLinkStatusExpired = "expired"
	PublicLinkStatusRevoked = "revoked"

	PublicLinkActorAnonymous    = "anonymous"
	PublicLinkActorExternal     = "authenticated_external"
	PublicLinkActorInternalUser = "internal_user"

	PublicLinkResultSuccess = "success"
	PublicLinkResultDenied  = "denied"

	PublicLinkActionOpenLink      = "open_link"
	PublicLinkActionPreview       = "preview"
	PublicLinkActionDownload      = "download"
	PublicLinkActionRevoke        = "revoke"
	PublicLinkActionUpdateSetting = "update_settings"

	PublicLinkDenyExpired       = "expired"
	PublicLinkDenyRevoked       = "revoked"
	PublicLinkDenyWrongPass     = "wrong_password"
	PublicLinkDenyPolicyBlocked = "policy_blocked"
)

var (
	ErrPublicLinkNotFound      = errors.New("public link not found")
	ErrPublicLinkExpired       = errors.New("public link expired")
	ErrPublicLinkRevoked       = errors.New("public link revoked")
	ErrPublicLinkAccessDenied  = errors.New("public link access denied")
	ErrPublicLinkWrongPassword = errors.New("public link wrong password")
	ErrSharedItemNotFound      = errors.New("shared item not found")
	ErrSharedItemNotFolder     = errors.New("shared item is not a folder")
	ErrSharedItemNotFile       = errors.New("shared item is not a file")
	ErrStorageNotConfigured    = errors.New("storage not configured")
	ErrFileHasNoStorageKey     = errors.New("file has no storage key")
)

type PublicLinkService struct {
	publicLinkRepo *repository.PublicLinkRepository
	itemRepo       *repository.ItemRepository
	b2Client       *storage.B2Client
	log            *zap.Logger
	sessionSecret  string
	sessionTTL     time.Duration

	dedupeMu     sync.Mutex
	accessDedup  map[string]time.Time
	dedupeWindow time.Duration
}

type PublicLinkAccessContext struct {
	Link          *model.PublicLink
	RootItem      *model.Item
	SessionClaims *jwtpkg.PublicLinkClaims
}

type PublicLinkStreamResult struct {
	Stream      *storage.FileStreamResult
	FileName    string
	ContentType string
}

func NewPublicLinkService(
	publicLinkRepo *repository.PublicLinkRepository,
	itemRepo *repository.ItemRepository,
	b2Client *storage.B2Client,
	sessionSecret string,
	log *zap.Logger,
) *PublicLinkService {
	return &PublicLinkService{
		publicLinkRepo: publicLinkRepo,
		itemRepo:       itemRepo,
		b2Client:       b2Client,
		sessionSecret:  sessionSecret,
		sessionTTL:     time.Hour,
		log:            log,
		accessDedup:    make(map[string]time.Time),
		dedupeWindow:   20 * time.Second,
	}
}

func (s *PublicLinkService) ListItemPublicLinks(ownerUserID, itemID uuid.UUID) ([]dto.PublicLinkResponse, error) {
	if _, err := s.itemRepo.FindByID(itemID, ownerUserID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSharedItemNotFound
		}
		return nil, fmt.Errorf("internal error")
	}

	links, err := s.publicLinkRepo.ListByItem(itemID, ownerUserID)
	if err != nil {
		s.log.Error("Failed to list public links", zap.Error(err), zap.String("item_id", itemID.String()))
		return nil, fmt.Errorf("internal error")
	}

	responses := make([]dto.PublicLinkResponse, 0, len(links))
	for i := range links {
		responses = append(responses, *ToPublicLinkResponse(&links[i], s.sessionTTL))
	}
	return responses, nil
}

func (s *PublicLinkService) CreatePublicLink(ownerUserID, itemID uuid.UUID, req *dto.CreatePublicLinkRequest) (*dto.PublicLinkResponse, error) {
	if _, err := s.itemRepo.FindByID(itemID, ownerUserID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSharedItemNotFound
		}
		return nil, fmt.Errorf("internal error")
	}

	expiresAt, err := parseExpiry(req.ExpiresAt)
	if err != nil {
		return nil, err
	}

	var passwordHash *string
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		hashed, err := hashpkg.HashPassword(*req.Password, hashpkg.DefaultParams())
		if err != nil {
			s.log.Error("Failed to hash public link password", zap.Error(err))
			return nil, fmt.Errorf("internal error")
		}
		passwordHash = &hashed
	}

	token, err := generatePublicLinkToken()
	if err != nil {
		s.log.Error("Failed to generate public link token", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}

	link := &model.PublicLink{
		OwnerUserID:    ownerUserID,
		ItemID:         itemID,
		Token:          token,
		PasswordHash:   passwordHash,
		ExpiresAt:      expiresAt,
		SessionVersion: 1,
	}

	if err := s.publicLinkRepo.Create(link); err != nil {
		s.log.Error("Failed to create public link", zap.Error(err), zap.String("item_id", itemID.String()))
		return nil, fmt.Errorf("internal error")
	}

	returnResponse := ToPublicLinkResponse(link, s.sessionTTL)
	return returnResponse, nil
}

func (s *PublicLinkService) UpdatePublicLink(ownerUserID, publicLinkID uuid.UUID, req *dto.UpdatePublicLinkRequest, ipAddress, userAgent string) (*dto.PublicLinkResponse, error) {
	link, err := s.publicLinkRepo.FindByID(publicLinkID, ownerUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPublicLinkNotFound
		}
		return nil, fmt.Errorf("internal error")
	}

	authChanged := false

	if req.PasswordEnabled != nil {
		if !*req.PasswordEnabled {
			if link.PasswordHash != nil {
				link.PasswordHash = nil
				authChanged = true
			}
		} else if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
			hashed, hashErr := hashpkg.HashPassword(*req.Password, hashpkg.DefaultParams())
			if hashErr != nil {
				s.log.Error("Failed to hash public link password", zap.Error(hashErr), zap.String("public_link_id", publicLinkID.String()))
				return nil, fmt.Errorf("internal error")
			}
			link.PasswordHash = &hashed
			authChanged = true
		} else if link.PasswordHash == nil {
			return nil, fmt.Errorf("password is required when enabling protection")
		}
	} else if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		hashed, hashErr := hashpkg.HashPassword(*req.Password, hashpkg.DefaultParams())
		if hashErr != nil {
			s.log.Error("Failed to hash public link password", zap.Error(hashErr), zap.String("public_link_id", publicLinkID.String()))
			return nil, fmt.Errorf("internal error")
		}
		link.PasswordHash = &hashed
		authChanged = true
	}

	if req.ClearExpiry {
		if link.ExpiresAt != nil {
			link.ExpiresAt = nil
			authChanged = true
		}
	} else if req.ExpiresAt != nil {
		expiresAt, parseErr := parseExpiry(req.ExpiresAt)
		if parseErr != nil {
			return nil, parseErr
		}
		if !sameTimePointer(link.ExpiresAt, expiresAt) {
			link.ExpiresAt = expiresAt
			authChanged = true
		}
	}

	if authChanged {
		link.SessionVersion++
	}

	if err := s.publicLinkRepo.Update(link); err != nil {
		s.log.Error("Failed to update public link", zap.Error(err), zap.String("public_link_id", publicLinkID.String()))
		return nil, fmt.Errorf("internal error")
	}

	_ = s.recordAudit(link, link.ItemID, nil, PublicLinkActorInternalUser, PublicLinkActionUpdateSetting, PublicLinkResultSuccess, nil, ipAddress, userAgent, false)

	return ToPublicLinkResponse(link, s.sessionTTL), nil
}

func (s *PublicLinkService) RevokePublicLink(ownerUserID, publicLinkID uuid.UUID, ipAddress, userAgent string) (*dto.PublicLinkResponse, error) {
	link, err := s.publicLinkRepo.FindByID(publicLinkID, ownerUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPublicLinkNotFound
		}
		return nil, fmt.Errorf("internal error")
	}

	if link.RevokedAt == nil {
		now := time.Now().UTC()
		link.RevokedAt = &now
		link.SessionVersion++
		if err := s.publicLinkRepo.Update(link); err != nil {
			s.log.Error("Failed to revoke public link", zap.Error(err), zap.String("public_link_id", publicLinkID.String()))
			return nil, fmt.Errorf("internal error")
		}
	}

	_ = s.recordAudit(link, link.ItemID, nil, PublicLinkActorInternalUser, PublicLinkActionRevoke, PublicLinkResultSuccess, nil, ipAddress, userAgent, false)

	return ToPublicLinkResponse(link, s.sessionTTL), nil
}

func (s *PublicLinkService) GetPublicLinkDetail(token, sessionToken, ipAddress, userAgent string) (*dto.PublicLinkDetailResponse, error) {
	accessContext, err := s.getAccessContext(token, sessionToken, false)
	if err != nil {
		if errors.Is(err, ErrPublicLinkExpired) {
			_ = s.recordDeniedByToken(token, PublicLinkActionOpenLink, PublicLinkDenyExpired, ipAddress, userAgent)
		}
		if errors.Is(err, ErrPublicLinkRevoked) {
			_ = s.recordDeniedByToken(token, PublicLinkActionOpenLink, PublicLinkDenyRevoked, ipAddress, userAgent)
		}
		return nil, err
	}

	var itemResponse *dto.PublicSharedItemResponse
	var sessionExpiresAt *string
	accessGranted := accessContext.Link.PasswordHash == nil

	if accessContext.SessionClaims != nil {
		accessGranted = true
		if accessContext.SessionClaims.ExpiresAt != nil {
			formatted := accessContext.SessionClaims.ExpiresAt.Time.UTC().Format(time.RFC3339)
			sessionExpiresAt = &formatted
		}
	}

	if accessGranted {
		rootItemResponse := ToPublicSharedItemResponse(accessContext.RootItem)
		itemResponse = &rootItemResponse
		dedupeKey := s.buildAccessDedupeKey(accessContext, PublicLinkActionOpenLink, accessContext.RootItem.ID, ipAddress, userAgent)
		_ = s.recordAudit(accessContext.Link, accessContext.RootItem.ID, &accessContext.RootItem.ID, PublicLinkActorAnonymous, PublicLinkActionOpenLink, PublicLinkResultSuccess, nil, ipAddress, userAgent, s.shouldRecordAccess(dedupeKey))
	}

	return &dto.PublicLinkDetailResponse{
		ID:                    accessContext.Link.ID.String(),
		Token:                 accessContext.Link.Token,
		Status:                publicLinkStatus(accessContext.Link, time.Now().UTC()),
		RequiresPassword:      accessContext.Link.PasswordHash != nil,
		AccessGranted:         accessGranted,
		ExternalAuthAvailable: false,
		ExpiresAt:             formatTimePointer(accessContext.Link.ExpiresAt),
		RevokedAt:             formatTimePointer(accessContext.Link.RevokedAt),
		SessionExpiresAt:      sessionExpiresAt,
		Item:                  itemResponse,
	}, nil
}

func (s *PublicLinkService) Authenticate(token, password, ipAddress, userAgent string) (*dto.AuthenticatePublicLinkResponse, error) {
	accessContext, err := s.getAccessContext(token, "", true)
	if err != nil {
		switch {
		case errors.Is(err, ErrPublicLinkExpired):
			_ = s.recordDeniedByToken(token, PublicLinkActionOpenLink, PublicLinkDenyExpired, ipAddress, userAgent)
		case errors.Is(err, ErrPublicLinkRevoked):
			_ = s.recordDeniedByToken(token, PublicLinkActionOpenLink, PublicLinkDenyRevoked, ipAddress, userAgent)
		}
		return nil, err
	}

	if accessContext.Link.PasswordHash == nil {
		return nil, ErrPublicLinkAccessDenied
	}

	matched, err := hashpkg.VerifyPassword(password, *accessContext.Link.PasswordHash)
	if err != nil {
		s.log.Error("Failed to verify public link password", zap.Error(err), zap.String("public_link_id", accessContext.Link.ID.String()))
		return nil, fmt.Errorf("internal error")
	}
	if !matched {
		_ = s.recordAudit(accessContext.Link, accessContext.RootItem.ID, nil, PublicLinkActorAnonymous, PublicLinkActionOpenLink, PublicLinkResultDenied, stringPointer(PublicLinkDenyWrongPass), ipAddress, userAgent, false)
		return nil, ErrPublicLinkWrongPassword
	}

	sessionToken, err := jwtpkg.GeneratePublicLinkSession(accessContext.Link.ID.String(), accessContext.Link.SessionVersion, s.sessionSecret, s.sessionTTL)
	if err != nil {
		s.log.Error("Failed to generate public link session", zap.Error(err), zap.String("public_link_id", accessContext.Link.ID.String()))
		return nil, fmt.Errorf("internal error")
	}

	expiresAt := time.Now().UTC().Add(s.sessionTTL)
	return &dto.AuthenticatePublicLinkResponse{
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
	}, nil
}

func (s *PublicLinkService) ListSharedItems(token, sessionToken string, parentID *uuid.UUID, ipAddress, userAgent string) (*dto.PublicFolderListingResponse, error) {
	accessContext, err := s.requireSharedAccess(token, sessionToken, PublicLinkActionOpenLink, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	if !accessContext.RootItem.IsFolder {
		return nil, ErrSharedItemNotFolder
	}

	currentFolder := accessContext.RootItem
	if parentID != nil {
		currentFolder, err = s.resolveTargetItem(accessContext.RootItem, accessContext.Link.OwnerUserID, *parentID)
		if err != nil {
			_ = s.recordAudit(accessContext.Link, accessContext.RootItem.ID, parentID, PublicLinkActorAnonymous, PublicLinkActionOpenLink, PublicLinkResultDenied, stringPointer(PublicLinkDenyPolicyBlocked), ipAddress, userAgent, false)
			return nil, ErrPublicLinkAccessDenied
		}
		if !currentFolder.IsFolder {
			return nil, ErrSharedItemNotFolder
		}
	}

	items, err := s.itemRepo.FindChildren(accessContext.Link.OwnerUserID, &currentFolder.ID)
	if err != nil {
		s.log.Error("Failed to list shared folder items", zap.Error(err), zap.String("public_link_id", accessContext.Link.ID.String()))
		return nil, fmt.Errorf("internal error")
	}

	responseItems := make([]dto.PublicSharedItemResponse, 0, len(items))
	for i := range items {
		if !isItemWithinScope(accessContext.RootItem, &items[i]) {
			continue
		}
		responseItems = append(responseItems, ToPublicSharedItemResponse(&items[i]))
	}

	breadcrumbs, err := s.buildBreadcrumbs(accessContext.RootItem, currentFolder, accessContext.Link.OwnerUserID)
	if err != nil {
		s.log.Warn("Failed to build public breadcrumbs", zap.Error(err), zap.String("public_link_id", accessContext.Link.ID.String()))
	}

	return &dto.PublicFolderListingResponse{
		RootItem:      ToPublicSharedItemResponse(accessContext.RootItem),
		CurrentFolder: ToPublicSharedItemResponse(currentFolder),
		Breadcrumbs:   breadcrumbs,
		Items:         responseItems,
	}, nil
}

func (s *PublicLinkService) StreamSharedItem(ctx context.Context, token, sessionToken string, itemID *uuid.UUID, download bool, rangeHeader, ipAddress, userAgent string) (*PublicLinkStreamResult, error) {
	action := PublicLinkActionPreview
	if download {
		action = PublicLinkActionDownload
	}

	accessContext, err := s.requireSharedAccess(token, sessionToken, action, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	targetItem := accessContext.RootItem
	if itemID != nil {
		targetItem, err = s.resolveTargetItem(accessContext.RootItem, accessContext.Link.OwnerUserID, *itemID)
		if err != nil {
			_ = s.recordAudit(accessContext.Link, accessContext.RootItem.ID, itemID, PublicLinkActorAnonymous, action, PublicLinkResultDenied, stringPointer(PublicLinkDenyPolicyBlocked), ipAddress, userAgent, false)
			return nil, ErrPublicLinkAccessDenied
		}
	}

	if targetItem.IsFolder {
		return nil, ErrSharedItemNotFile
	}
	if targetItem.StorageKey == nil {
		return nil, ErrFileHasNoStorageKey
	}
	if s.b2Client == nil {
		return nil, ErrStorageNotConfigured
	}

	dedupeKey := s.buildAccessDedupeKey(accessContext, action, targetItem.ID, ipAddress, userAgent)
	shouldCount := s.shouldRecordAccess(dedupeKey)
	_ = s.recordAudit(accessContext.Link, accessContext.RootItem.ID, &targetItem.ID, PublicLinkActorAnonymous, action, PublicLinkResultSuccess, nil, ipAddress, userAgent, shouldCount)

	streamResult, err := s.b2Client.GetFileStream(ctx, *targetItem.StorageKey, rangeHeader)
	if err != nil {
		s.log.Error("Failed to get public link file stream", zap.Error(err), zap.String("public_link_id", accessContext.Link.ID.String()), zap.String("item_id", targetItem.ID.String()))
		return nil, fmt.Errorf("failed to stream file")
	}

	contentType := streamResult.ContentType
	if targetItem.MimeType != nil && *targetItem.MimeType != "" {
		contentType = *targetItem.MimeType
	}

	return &PublicLinkStreamResult{
		Stream:      streamResult,
		FileName:    targetItem.Name,
		ContentType: contentType,
	}, nil
}

func (s *PublicLinkService) BuildContentDisposition(fileName string, download bool) string {
	dispositionType := "inline"
	if download {
		dispositionType = "attachment"
	}
	return mime.FormatMediaType(dispositionType, map[string]string{"filename": fileName})
}

func (s *PublicLinkService) requireSharedAccess(token, sessionToken, action, ipAddress, userAgent string) (*PublicLinkAccessContext, error) {
	accessContext, err := s.getAccessContext(token, sessionToken, false)
	if err != nil {
		switch {
		case errors.Is(err, ErrPublicLinkExpired):
			_ = s.recordDeniedByToken(token, action, PublicLinkDenyExpired, ipAddress, userAgent)
			return nil, err
		case errors.Is(err, ErrPublicLinkRevoked):
			_ = s.recordDeniedByToken(token, action, PublicLinkDenyRevoked, ipAddress, userAgent)
			return nil, err
		default:
			_ = s.recordDeniedByToken(token, action, PublicLinkDenyPolicyBlocked, ipAddress, userAgent)
			return nil, ErrPublicLinkAccessDenied
		}
	}

	if accessContext.Link.PasswordHash != nil && accessContext.SessionClaims == nil {
		_ = s.recordAudit(accessContext.Link, accessContext.RootItem.ID, nil, PublicLinkActorAnonymous, action, PublicLinkResultDenied, stringPointer(PublicLinkDenyPolicyBlocked), ipAddress, userAgent, false)
		return nil, ErrPublicLinkAccessDenied
	}

	return accessContext, nil
}

func (s *PublicLinkService) getAccessContext(token, sessionToken string, skipSessionValidation bool) (*PublicLinkAccessContext, error) {
	link, err := s.publicLinkRepo.FindByToken(token)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPublicLinkNotFound
		}
		return nil, fmt.Errorf("internal error")
	}

	status := publicLinkStatus(link, time.Now().UTC())
	switch status {
	case PublicLinkStatusExpired:
		return nil, ErrPublicLinkExpired
	case PublicLinkStatusRevoked:
		return nil, ErrPublicLinkRevoked
	}

	rootItem, err := s.itemRepo.FindByID(link.ItemID, link.OwnerUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSharedItemNotFound
		}
		return nil, fmt.Errorf("internal error")
	}

	accessContext := &PublicLinkAccessContext{
		Link:     link,
		RootItem: rootItem,
	}

	if skipSessionValidation || link.PasswordHash == nil || strings.TrimSpace(sessionToken) == "" {
		return accessContext, nil
	}

	claims, err := jwtpkg.ValidatePublicLinkSession(sessionToken, s.sessionSecret)
	if err != nil {
		return accessContext, nil
	}
	if claims.LinkID != link.ID.String() || claims.SessionVersion != link.SessionVersion {
		return accessContext, nil
	}

	accessContext.SessionClaims = claims
	return accessContext, nil
}

func (s *PublicLinkService) resolveTargetItem(rootItem *model.Item, ownerUserID, itemID uuid.UUID) (*model.Item, error) {
	if itemID == rootItem.ID {
		return rootItem, nil
	}

	item, err := s.itemRepo.FindByID(itemID, ownerUserID)
	if err != nil {
		return nil, err
	}
	if !isItemWithinScope(rootItem, item) {
		return nil, ErrPublicLinkAccessDenied
	}
	return item, nil
}

func (s *PublicLinkService) buildBreadcrumbs(rootItem, currentFolder *model.Item, ownerUserID uuid.UUID) ([]dto.PublicSharedItemResponse, error) {
	if currentFolder == nil {
		return nil, nil
	}

	chain := []*model.Item{currentFolder}
	cursor := currentFolder

	for cursor.ID != rootItem.ID {
		if cursor.ParentID == nil {
			break
		}
		parent, err := s.itemRepo.FindByID(*cursor.ParentID, ownerUserID)
		if err != nil {
			return nil, err
		}
		if !isItemWithinScope(rootItem, parent) {
			break
		}
		chain = append(chain, parent)
		cursor = parent
	}

	breadcrumbs := make([]dto.PublicSharedItemResponse, 0, len(chain))
	for i := len(chain) - 1; i >= 0; i-- {
		breadcrumbs = append(breadcrumbs, ToPublicSharedItemResponse(chain[i]))
	}

	return breadcrumbs, nil
}

func (s *PublicLinkService) recordDeniedByToken(token, action, denyReason, ipAddress, userAgent string) error {
	link, err := s.publicLinkRepo.FindByToken(token)
	if err != nil {
		return err
	}

	itemID := link.ItemID
	return s.recordAudit(link, itemID, nil, PublicLinkActorAnonymous, action, PublicLinkResultDenied, &denyReason, ipAddress, userAgent, false)
}

func (s *PublicLinkService) recordAudit(link *model.PublicLink, itemID uuid.UUID, requestedItemID *uuid.UUID, actorType, action, result string, denyReason *string, ipAddress, userAgent string, countAccess bool) error {
	entry := &model.PublicLinkAuditLog{
		PublicLinkID:    link.ID,
		ItemID:          itemID,
		RequestedItemID: requestedItemID,
		ActorType:       actorType,
		Action:          action,
		Result:          result,
		DenyReason:      denyReason,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		CreatedAt:       time.Now().UTC(),
	}

	if err := s.publicLinkRepo.CreateAuditLog(entry); err != nil {
		s.log.Warn("Failed to write public link audit log", zap.Error(err), zap.String("public_link_id", link.ID.String()))
		return err
	}

	if countAccess {
		if err := s.publicLinkRepo.IncrementAccess(link.ID, entry.CreatedAt); err != nil {
			s.log.Warn("Failed to update public link access stats", zap.Error(err), zap.String("public_link_id", link.ID.String()))
			return err
		}
		link.AccessCount++
		link.LastAccessedAt = &entry.CreatedAt
	}

	return nil
}

func (s *PublicLinkService) shouldRecordAccess(key string) bool {
	if key == "" {
		return true
	}

	now := time.Now().UTC()

	s.dedupeMu.Lock()
	defer s.dedupeMu.Unlock()

	for existingKey, timestamp := range s.accessDedup {
		if now.Sub(timestamp) > s.dedupeWindow {
			delete(s.accessDedup, existingKey)
		}
	}

	lastSeen, ok := s.accessDedup[key]
	if ok && now.Sub(lastSeen) <= s.dedupeWindow {
		return false
	}

	s.accessDedup[key] = now
	return true
}

func (s *PublicLinkService) buildAccessDedupeKey(accessContext *PublicLinkAccessContext, action string, itemID uuid.UUID, ipAddress, userAgent string) string {
	sessionKey := ipAddress + ":" + userAgent
	if accessContext.SessionClaims != nil {
		sessionKey = accessContext.SessionClaims.ID
	}
	return strings.Join([]string{
		accessContext.Link.ID.String(),
		action,
		itemID.String(),
		sessionKey,
	}, ":")
}

func ToPublicLinkResponse(link *model.PublicLink, sessionTTL time.Duration) *dto.PublicLinkResponse {
	return &dto.PublicLinkResponse{
		ID:                link.ID.String(),
		ItemID:            link.ItemID.String(),
		Token:             link.Token,
		Status:            publicLinkStatus(link, time.Now().UTC()),
		RequiresPassword:  link.PasswordHash != nil,
		ExpiresAt:         formatTimePointer(link.ExpiresAt),
		RevokedAt:         formatTimePointer(link.RevokedAt),
		AccessCount:       link.AccessCount,
		LastAccessedAt:    formatTimePointer(link.LastAccessedAt),
		SessionTTLSeconds: int64(sessionTTL.Seconds()),
		CreatedAt:         link.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         link.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func ToPublicSharedItemResponse(item *model.Item) dto.PublicSharedItemResponse {
	return dto.PublicSharedItemResponse{
		ID:        item.ID.String(),
		Name:      item.Name,
		IsFolder:  item.IsFolder,
		MimeType:  item.MimeType,
		Size:      item.Size,
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func publicLinkStatus(link *model.PublicLink, now time.Time) string {
	switch {
	case link.RevokedAt != nil:
		return PublicLinkStatusRevoked
	case link.ExpiresAt != nil && now.After(link.ExpiresAt.UTC()):
		return PublicLinkStatusExpired
	default:
		return PublicLinkStatusActive
	}
}

func parseExpiry(raw *string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return nil, fmt.Errorf("invalid expires_at")
	}
	parsed = parsed.UTC()
	if !parsed.After(time.Now().UTC()) {
		return nil, fmt.Errorf("expires_at must be in the future")
	}
	return &parsed, nil
}

func generatePublicLinkToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func formatTimePointer(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

func sameTimePointer(left, right *time.Time) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return left.UTC().Equal(right.UTC())
}

func isItemWithinScope(rootItem, candidate *model.Item) bool {
	if rootItem == nil || candidate == nil {
		return false
	}
	if rootItem.ID == candidate.ID {
		return true
	}
	if !rootItem.IsFolder {
		return false
	}
	return strings.HasPrefix(candidate.Path, rootItem.Path+"/")
}

func stringPointer(value string) *string {
	return &value
}
