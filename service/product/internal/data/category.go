package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/storm/myidea/service/product/internal/biz"
)

type GORMCategory struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	ParentID  int64     `gorm:"index;not null;default:0"`
	Name      string    `gorm:"size:64;not null"`
	Icon      string    `gorm:"size:256"`
	SortOrder int32     `gorm:"default:0"`
	Level     int32     `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (GORMCategory) TableName() string { return "categories" }

type categoryRepo struct {
	data *Data
}

func NewCategoryRepo(data *Data) biz.CategoryRepository {
	return &categoryRepo{data: data}
}

func (r *categoryRepo) Create(ctx context.Context, c *biz.Category) (int64, error) {
	po := &GORMCategory{
		ParentID:  c.ParentID,
		Name:      c.Name,
		Icon:      c.Icon,
		SortOrder: c.SortOrder,
		Level:     c.Level,
	}
	if err := r.data.DB(ctx).Create(po).Error; err != nil {
		return 0, fmt.Errorf("create category: %w", err)
	}
	return po.ID, nil
}

func (r *categoryRepo) Update(ctx context.Context, c *biz.Category) error {
	po := &GORMCategory{
		ID:        c.ID,
		ParentID:  c.ParentID,
		Name:      c.Name,
		Icon:      c.Icon,
		SortOrder: c.SortOrder,
		Level:     c.Level,
	}
	return r.data.DB(ctx).Model(&GORMCategory{}).Where("id = ?", c.ID).Updates(map[string]any{
		"name":       po.Name,
		"icon":       po.Icon,
		"sort_order": po.SortOrder,
	}).Error
}

func (r *categoryRepo) Delete(ctx context.Context, id int64) error {
	res := r.data.DB(ctx).Delete(&GORMCategory{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete category: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return biz.ErrCategoryNotFound
	}
	return nil
}

func (r *categoryRepo) GetByID(ctx context.Context, id int64) (*biz.Category, error) {
	var po GORMCategory
	err := r.data.DB(ctx).First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by id: %w", err)
	}
	return r.toDomain(&po), nil
}

func (r *categoryRepo) GetTree(ctx context.Context, parentID int64) ([]*biz.Category, error) {
	var pos []GORMCategory
	err := r.data.DB(ctx).Where("parent_id = ?", parentID).Order("sort_order ASC, id ASC").Find(&pos).Error
	if err != nil {
		return nil, fmt.Errorf("get category tree: %w", err)
	}
	categories := make([]*biz.Category, 0, len(pos))
	for i := range pos {
		categories = append(categories, r.toDomain(&pos[i]))
	}
	return categories, nil
}

func (r *categoryRepo) HasChildren(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.data.DB(ctx).Model(&GORMCategory{}).Where("parent_id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("count children: %w", err)
	}
	return count > 0, nil
}

func (r *categoryRepo) toDomain(po *GORMCategory) *biz.Category {
	return &biz.Category{
		ID:        po.ID,
		ParentID:  po.ParentID,
		Name:      po.Name,
		Icon:      po.Icon,
		SortOrder: po.SortOrder,
		Level:     po.Level,
		CreatedAt: po.CreatedAt,
	}
}
