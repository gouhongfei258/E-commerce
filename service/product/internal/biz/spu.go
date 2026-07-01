package biz

import (
	"context"
	"fmt"
	"time"
)

type SPUStatus int32

const (
	SPUStatusOffline SPUStatus = 0
	SPUStatusOnline  SPUStatus = 1
	SPUStatusSoldOut SPUStatus = 2
)

type SPU struct {
	ID               int64
	CategoryID       int64
	BrandID          int64
	Title            string
	Subtitle         string
	Status           SPUStatus
	SaleableAttrNames []string
	SaleCount        int32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (s SPUStatus) CanTransitionTo(target SPUStatus) bool {
	switch s {
	case SPUStatusOffline:
		return target == SPUStatusOnline
	case SPUStatusOnline:
		return target == SPUStatusOffline || target == SPUStatusSoldOut
	case SPUStatusSoldOut:
		return target == SPUStatusOnline
	default:
		return false
	}
}

type SPUFilter struct {
	CategoryID int64
	BrandID    int64
	Keyword    string
	Status     SPUStatus
}

type SPURepository interface {
	Create(ctx context.Context, s *SPU) (int64, error)
	Update(ctx context.Context, s *SPU) error
	GetByID(ctx context.Context, id int64) (*SPU, error)
	List(ctx context.Context, filter *SPUFilter, page, pageSize int32) ([]*SPU, int32, error)
}

type SPUBiz struct {
	repo SPURepository
}

func NewSPUBiz(repo SPURepository) *SPUBiz {
	return &SPUBiz{repo: repo}
}

func (b *SPUBiz) CreateSPU(ctx context.Context, categoryID, brandID int64, title, subtitle string, saleableAttrNames []string) (*SPU, error) {
	spu := &SPU{
		CategoryID:       categoryID,
		BrandID:          brandID,
		Title:            title,
		Subtitle:         subtitle,
		Status:           SPUStatusOffline,
		SaleableAttrNames: saleableAttrNames,
	}

	id, err := b.repo.Create(ctx, spu)
	if err != nil {
		return nil, fmt.Errorf("create spu: %w", err)
	}
	spu.ID = id
	return spu, nil
}

func (b *SPUBiz) UpdateSPU(ctx context.Context, spu *SPU) (*SPU, error) {
	existing, err := b.repo.GetByID(ctx, spu.ID)
	if err != nil {
		return nil, fmt.Errorf("get spu: %w", err)
	}

	existing.CategoryID = spu.CategoryID
	existing.BrandID = spu.BrandID
	existing.Title = spu.Title
	existing.Subtitle = spu.Subtitle
	existing.SaleableAttrNames = spu.SaleableAttrNames

	if spu.Status != existing.Status && spu.Status != 0 {
		if !existing.Status.CanTransitionTo(spu.Status) {
			return nil, ErrSPUStatusInvalid
		}
		existing.Status = spu.Status
	}

	if err := b.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update spu: %w", err)
	}
	return existing, nil
}

func (b *SPUBiz) GetSPU(ctx context.Context, id int64) (*SPU, error) {
	return b.repo.GetByID(ctx, id)
}

func (b *SPUBiz) ListSPUs(ctx context.Context, filter *SPUFilter, page, pageSize int32) ([]*SPU, int32, error) {
	return b.repo.List(ctx, filter, page, pageSize)
}
