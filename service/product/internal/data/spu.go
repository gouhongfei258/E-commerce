package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/storm/myidea/service/product/internal/biz"
)

type GORMSPU struct {
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	CategoryID       int64     `gorm:"index;not null"`
	BrandID          int64     `gorm:"index;not null"`
	Title            string    `gorm:"size:256;not null"`
	Subtitle         string    `gorm:"size:512"`
	Status           int32     `gorm:"default:0;not null"`
	SaleableAttrNames string   `gorm:"size:256"`
	SaleCount        int32     `gorm:"default:0"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

func (GORMSPU) TableName() string { return "spus" }

func (s *GORMSPU) BeforeCreate(tx *gorm.DB) error {
	if s.SaleableAttrNames == "" {
		s.SaleableAttrNames = "[]"
	}
	return nil
}

type spuRepo struct {
	data *Data
}

func NewSPURepo(data *Data) biz.SPURepository {
	return &spuRepo{data: data}
}

func (r *spuRepo) Create(ctx context.Context, s *biz.SPU) (int64, error) {
	attrsJSON, _ := json.Marshal(s.SaleableAttrNames)
	po := &GORMSPU{
		CategoryID:       s.CategoryID,
		BrandID:          s.BrandID,
		Title:            s.Title,
		Subtitle:         s.Subtitle,
		Status:           int32(s.Status),
		SaleableAttrNames: string(attrsJSON),
	}
	if err := r.data.DB(ctx).Create(po).Error; err != nil {
		return 0, fmt.Errorf("create spu: %w", err)
	}
	return po.ID, nil
}

func (r *spuRepo) Update(ctx context.Context, s *biz.SPU) error {
	attrsJSON, _ := json.Marshal(s.SaleableAttrNames)
	updates := map[string]any{
		"category_id":        s.CategoryID,
		"brand_id":           s.BrandID,
		"title":              s.Title,
		"subtitle":           s.Subtitle,
		"status":             s.Status,
		"saleable_attr_names": string(attrsJSON),
	}
	return r.data.DB(ctx).Model(&GORMSPU{}).Where("id = ?", s.ID).Updates(updates).Error
}

func (r *spuRepo) GetByID(ctx context.Context, id int64) (*biz.SPU, error) {
	var po GORMSPU
	err := r.data.DB(ctx).First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrSPUNotFound
		}
		return nil, fmt.Errorf("get spu by id: %w", err)
	}
	return r.toDomain(&po), nil
}

func (r *spuRepo) List(ctx context.Context, filter *biz.SPUFilter, page, pageSize int32) ([]*biz.SPU, int32, error) {
	db := r.data.DB(ctx).Model(&GORMSPU{})
	if filter != nil {
		if filter.CategoryID > 0 {
			db = db.Where("category_id = ?", filter.CategoryID)
		}
		if filter.BrandID > 0 {
			db = db.Where("brand_id = ?", filter.BrandID)
		}
		if filter.Keyword != "" {
			db = db.Where("title LIKE ? OR subtitle LIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
		}
		if filter.Status > 0 {
			db = db.Where("status = ?", filter.Status)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count spus: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var pos []GORMSPU
	if err := db.Order("id DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("list spus: %w", err)
	}

	spus := make([]*biz.SPU, 0, len(pos))
	for i := range pos {
		spus = append(spus, r.toDomain(&pos[i]))
	}
	return spus, int32(total), nil
}

func (r *spuRepo) toDomain(po *GORMSPU) *biz.SPU {
	var attrs []string
	json.Unmarshal([]byte(po.SaleableAttrNames), &attrs)
	if attrs == nil {
		attrs = []string{}
	}
	return &biz.SPU{
		ID:               po.ID,
		CategoryID:       po.CategoryID,
		BrandID:          po.BrandID,
		Title:            po.Title,
		Subtitle:         po.Subtitle,
		Status:           biz.SPUStatus(po.Status),
		SaleableAttrNames: attrs,
		SaleCount:        po.SaleCount,
		CreatedAt:        po.CreatedAt,
		UpdatedAt:        po.UpdatedAt,
	}
}
