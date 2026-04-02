package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/biho/onedrive/internal/model"
)

type PublicLinkRepository struct {
	db *gorm.DB
}

func NewPublicLinkRepository(db *gorm.DB) *PublicLinkRepository {
	return &PublicLinkRepository{db: db}
}

func (r *PublicLinkRepository) Create(link *model.PublicLink) error {
	return r.db.Create(link).Error
}

func (r *PublicLinkRepository) FindByID(id, ownerUserID uuid.UUID) (*model.PublicLink, error) {
	var link model.PublicLink
	err := r.db.
		Where("id = ? AND owner_user_id = ?", id, ownerUserID).
		First(&link).Error
	return &link, err
}

func (r *PublicLinkRepository) FindByToken(token string) (*model.PublicLink, error) {
	var link model.PublicLink
	err := r.db.
		Where("token = ?", token).
		First(&link).Error
	return &link, err
}

func (r *PublicLinkRepository) ListByItem(itemID, ownerUserID uuid.UUID) ([]model.PublicLink, error) {
	var links []model.PublicLink
	err := r.db.
		Where("item_id = ? AND owner_user_id = ?", itemID, ownerUserID).
		Order("created_at DESC").
		Find(&links).Error
	return links, err
}

func (r *PublicLinkRepository) Update(link *model.PublicLink) error {
	return r.db.Save(link).Error
}

func (r *PublicLinkRepository) IncrementAccess(linkID uuid.UUID, accessedAt time.Time) error {
	return r.db.Model(&model.PublicLink{}).
		Where("id = ?", linkID).
		Updates(map[string]interface{}{
			"access_count":     gorm.Expr("access_count + ?", 1),
			"last_accessed_at": accessedAt,
		}).Error
}

func (r *PublicLinkRepository) CreateAuditLog(entry *model.PublicLinkAuditLog) error {
	return r.db.Create(entry).Error
}

func (r *PublicLinkRepository) UpsertAuditLog(entry *model.PublicLinkAuditLog, conflictColumns ...string) error {
	if len(conflictColumns) == 0 {
		return r.db.Create(entry).Error
	}

	columns := make([]clause.Column, 0, len(conflictColumns))
	for _, column := range conflictColumns {
		columns = append(columns, clause.Column{Name: column})
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   columns,
		DoNothing: true,
	}).Create(entry).Error
}
