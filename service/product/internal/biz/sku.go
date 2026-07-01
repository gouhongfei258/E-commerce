package biz

import (
	"context"
	"fmt"
	"time"
)

type SKU struct {
	ID          int64
	SPUID       int64
	Attrs       map[string]string
	Price       float64
	OriginPrice float64
	Stock       int32
	LockedStock int32
	Code        string
	Image       string
	Status      int32
	SaleCount   int32
	Version     int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *SKU) AvailableStock() int32 {
	return s.Stock - s.LockedStock
}

type SKURepository interface {
	BatchCreate(ctx context.Context, skus []*SKU) error
	Update(ctx context.Context, sku *SKU) error
	ListBySPUID(ctx context.Context, spuID int64) ([]*SKU, error)
	GetByID(ctx context.Context, id int64) (*SKU, error)
	LockStock(ctx context.Context, id int64, quantity int32, version int32) (bool, error)
	ConfirmDeduct(ctx context.Context, id int64, quantity int32) error
	UnlockStock(ctx context.Context, id int64, quantity int32) error
	CreateJournal(ctx context.Context, skuID int64, orderNo string, changeType int32, quantity int32) error
}

type SKUBiz struct {
	repo SKURepository
}

func NewSKUBiz(repo SKURepository) *SKUBiz {
	return &SKUBiz{repo: repo}
}

func (b *SKUBiz) BatchCreateSKU(ctx context.Context, spuID int64, skus []*SKU) ([]*SKU, error) {
	for i := range skus {
		skus[i].SPUID = spuID
		skus[i].Status = 1
		skus[i].Version = 1
	}

	if err := b.repo.BatchCreate(ctx, skus); err != nil {
		return nil, fmt.Errorf("batch create skus: %w", err)
	}
	return skus, nil
}

func (b *SKUBiz) UpdateSKU(ctx context.Context, sku *SKU) error {
	existing, err := b.repo.GetByID(ctx, sku.ID)
	if err != nil {
		return fmt.Errorf("get sku: %w", err)
	}

	existing.Price = sku.Price
	existing.OriginPrice = sku.OriginPrice
	existing.Code = sku.Code
	existing.Image = sku.Image
	existing.Status = sku.Status

	return b.repo.Update(ctx, existing)
}

func (b *SKUBiz) ListSKUs(ctx context.Context, spuID int64) ([]*SKU, error) {
	return b.repo.ListBySPUID(ctx, spuID)
}

func (b *SKUBiz) LockStock(ctx context.Context, skuID int64, quantity int32, orderNo string) error {
	if err := b.repo.CreateJournal(ctx, skuID, orderNo, 1, quantity); err != nil {
		return ErrStockOpDuplicated
	}

	sku, err := b.repo.GetByID(ctx, skuID)
	if err != nil {
		return fmt.Errorf("get sku for lock: %w", err)
	}

	ok, err := b.repo.LockStock(ctx, skuID, quantity, sku.Version)
	if err != nil {
		return fmt.Errorf("lock stock: %w", err)
	}
	if !ok {
		return ErrStockInsufficient
	}
	return nil
}

func (b *SKUBiz) ConfirmDeductStock(ctx context.Context, skuID int64, quantity int32, orderNo string) error {
	if err := b.repo.CreateJournal(ctx, skuID, orderNo, 2, quantity); err != nil {
		return ErrStockOpDuplicated
	}

	if err := b.repo.ConfirmDeduct(ctx, skuID, quantity); err != nil {
		return fmt.Errorf("confirm deduct: %w", err)
	}
	return nil
}

func (b *SKUBiz) UnlockStock(ctx context.Context, skuID int64, quantity int32, orderNo string) error {
	if err := b.repo.CreateJournal(ctx, skuID, orderNo, 3, quantity); err != nil {
		return ErrStockOpDuplicated
	}

	if err := b.repo.UnlockStock(ctx, skuID, quantity); err != nil {
		return fmt.Errorf("unlock stock: %w", err)
	}
	return nil
}

const (
	StockChangeLock    int32 = 1
	StockChangeConfirm int32 = 2
	StockChangeUnlock  int32 = 3
)
