package data

import (
	"context"
	"fmt"
	"time"

	"github.com/storm/myidea/service/product/internal/biz"
)

type GORMBrand struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"uniqueIndex;size:128;not null"`
	Logo      string    `gorm:"size:256"`
	SortOrder int32     `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (GORMBrand) TableName() string { return "brands" }

type brandRepo struct {
	data *Data
}

func NewBrandRepo(data *Data) biz.BrandRepository {
	return &brandRepo{data: data}
}

func (r *brandRepo) Create(ctx context.Context, b *biz.Brand) (int64, error) {
	po := &GORMBrand{
		Name:      b.Name,
		Logo:      b.Logo,
		SortOrder: b.SortOrder,
	}
	if err := r.data.DB(ctx).Create(po).Error; err != nil {
		return 0, fmt.Errorf("create brand: %w", err)
	}
	return po.ID, nil
}

func (r *brandRepo) List(ctx context.Context, keyword string, page, pageSize int32) ([]*biz.Brand, int32, error) {
	db := r.data.DB(ctx).Model(&GORMBrand{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count brands: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var pos []GORMBrand
	if err := db.Order("sort_order ASC, id DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("list brands: %w", err)
	}

	brands := make([]*biz.Brand, 0, len(pos))
	for i := range pos {
		brands = append(brands, &biz.Brand{
			ID:        pos[i].ID,
			Name:      pos[i].Name,
			Logo:      pos[i].Logo,
			SortOrder: pos[i].SortOrder,
			CreatedAt: pos[i].CreatedAt,
		})
	}
	return brands, int32(total), nil
}
