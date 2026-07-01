package biz

import (
	"context"
	"fmt"
	"time"
)

type Brand struct {
	ID        int64
	Name      string
	Logo      string
	SortOrder int32
	CreatedAt time.Time
}

type BrandRepository interface {
	Create(ctx context.Context, b *Brand) (int64, error)
	List(ctx context.Context, keyword string, page, pageSize int32) ([]*Brand, int32, error)
}

type BrandBiz struct {
	repo BrandRepository
}

func NewBrandBiz(repo BrandRepository) *BrandBiz {
	return &BrandBiz{repo: repo}
}

func (b *BrandBiz) CreateBrand(ctx context.Context, name, logo string, sortOrder int32) (*Brand, error) {
	brand := &Brand{
		Name:      name,
		Logo:      logo,
		SortOrder: sortOrder,
	}

	id, err := b.repo.Create(ctx, brand)
	if err != nil {
		return nil, fmt.Errorf("create brand: %w", err)
	}
	brand.ID = id
	return brand, nil
}

func (b *BrandBiz) ListBrands(ctx context.Context, keyword string, page, pageSize int32) ([]*Brand, int32, error) {
	return b.repo.List(ctx, keyword, page, pageSize)
}
