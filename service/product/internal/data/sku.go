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

type GORMSKU struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	SPUID       int64     `gorm:"index;not null"`
	Attrs       string    `gorm:"type:json"`
	Price       float64   `gorm:"type:decimal(10,2);not null"`
	OriginPrice float64   `gorm:"type:decimal(10,2);default:0"`
	Stock       int32     `gorm:"default:0"`
	LockedStock int32     `gorm:"default:0"`
	Code        string    `gorm:"size:64"`
	Image       string    `gorm:"size:256"`
	Status      int32     `gorm:"default:1"`
	SaleCount   int32     `gorm:"default:0"`
	Version     int32     `gorm:"default:1"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (GORMSKU) TableName() string { return "skus" }

type GORMStockJournal struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	SKUID      int64     `gorm:"index;not null;uniqueIndex:uk_order_sku_type"`
	OrderNo    string    `gorm:"size:32;not null;uniqueIndex:uk_order_sku_type"`
	ChangeType int32     `gorm:"not null;uniqueIndex:uk_order_sku_type"`
	Quantity   int32     `gorm:"not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (GORMStockJournal) TableName() string { return "stock_journals" }

type skuRepo struct {
	data *Data
}

func NewSKURepo(data *Data) biz.SKURepository {
	return &skuRepo{data: data}
}

func (r *skuRepo) BatchCreate(ctx context.Context, skus []*biz.SKU) error {
	pos := make([]GORMSKU, len(skus))
	for i, s := range skus {
		attrsJSON, _ := json.Marshal(s.Attrs)
		pos[i] = GORMSKU{
			SPUID:       s.SPUID,
			Attrs:       string(attrsJSON),
			Price:       s.Price,
			OriginPrice: s.OriginPrice,
			Stock:       s.Stock,
			Code:        s.Code,
			Image:       s.Image,
			Status:      1,
			Version:     1,
		}
	}
	if err := r.data.DB(ctx).Create(&pos).Error; err != nil {
		return fmt.Errorf("batch create skus: %w", err)
	}
	for i := range skus {
		skus[i].ID = pos[i].ID
	}
	return nil
}

func (r *skuRepo) Update(ctx context.Context, sku *biz.SKU) error {
	return r.data.DB(ctx).Model(&GORMSKU{}).Where("id = ?", sku.ID).Updates(map[string]any{
		"price":        sku.Price,
		"origin_price": sku.OriginPrice,
		"code":         sku.Code,
		"image":        sku.Image,
		"status":       sku.Status,
	}).Error
}

func (r *skuRepo) ListBySPUID(ctx context.Context, spuID int64) ([]*biz.SKU, error) {
	var pos []GORMSKU
	err := r.data.DB(ctx).Where("spu_id = ?", spuID).Order("id ASC").Find(&pos).Error
	if err != nil {
		return nil, fmt.Errorf("list skus by spu id: %w", err)
	}
	skus := make([]*biz.SKU, 0, len(pos))
	for i := range pos {
		skus = append(skus, r.toDomain(&pos[i]))
	}
	return skus, nil
}

func (r *skuRepo) GetByID(ctx context.Context, id int64) (*biz.SKU, error) {
	var po GORMSKU
	err := r.data.DB(ctx).First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrSKUNotFound
		}
		return nil, fmt.Errorf("get sku by id: %w", err)
	}
	return r.toDomain(&po), nil
}

func (r *skuRepo) LockStock(ctx context.Context, id int64, quantity int32, version int32) (bool, error) {
	res := r.data.DB(ctx).Model(&GORMSKU{}).
		Where("id = ? AND version = ? AND stock >= ?", id, version, quantity).
		Updates(map[string]any{
			"stock":        gorm.Expr("stock - ?", quantity),
			"locked_stock": gorm.Expr("locked_stock + ?", quantity),
			"version":      gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return false, fmt.Errorf("lock stock: %w", res.Error)
	}
	return res.RowsAffected > 0, nil
}

func (r *skuRepo) ConfirmDeduct(ctx context.Context, id int64, quantity int32) error {
	res := r.data.DB(ctx).Model(&GORMSKU{}).
		Where("id = ? AND locked_stock >= ?", id, quantity).
		Update("locked_stock", gorm.Expr("locked_stock - ?", quantity))
	if res.Error != nil {
		return fmt.Errorf("confirm deduct: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return biz.ErrStockInsufficient
	}
	return nil
}

func (r *skuRepo) UnlockStock(ctx context.Context, id int64, quantity int32) error {
	res := r.data.DB(ctx).Model(&GORMSKU{}).
		Where("id = ? AND locked_stock >= ?", id, quantity).
		Updates(map[string]any{
			"stock":        gorm.Expr("stock + ?", quantity),
			"locked_stock": gorm.Expr("locked_stock - ?", quantity),
			"version":      gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return fmt.Errorf("unlock stock: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return biz.ErrStockInsufficient
	}
	return nil
}

func (r *skuRepo) CreateJournal(ctx context.Context, skuID int64, orderNo string, changeType int32, quantity int32) error {
	var existing GORMStockJournal
	err := r.data.DB(ctx).
		Where("order_no = ? AND sku_id = ? AND change_type = ?", orderNo, skuID, changeType).
		First(&existing).Error
	if err == nil {
		return ErrJournalAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("check journal: %w", err)
	}

	po := &GORMStockJournal{
		SKUID:      skuID,
		OrderNo:    orderNo,
		ChangeType: changeType,
		Quantity:   quantity,
	}
	if err := r.data.DB(ctx).Create(po).Error; err != nil {
		return ErrJournalAlreadyExists
	}
	return nil
}

var ErrJournalAlreadyExists = fmt.Errorf("stock journal entry already exists")

func (r *skuRepo) toDomain(po *GORMSKU) *biz.SKU {
	var attrs map[string]string
	json.Unmarshal([]byte(po.Attrs), &attrs)
	if attrs == nil {
		attrs = make(map[string]string)
	}
	return &biz.SKU{
		ID:          po.ID,
		SPUID:       po.SPUID,
		Attrs:       attrs,
		Price:       po.Price,
		OriginPrice: po.OriginPrice,
		Stock:       po.Stock,
		LockedStock: po.LockedStock,
		Code:        po.Code,
		Image:       po.Image,
		Status:      po.Status,
		SaleCount:   po.SaleCount,
		Version:     po.Version,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}
