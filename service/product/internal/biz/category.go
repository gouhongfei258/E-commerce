package biz

import (
	"context"
	"fmt"
	"time"
)

type Category struct {
	ID        int64
	ParentID  int64
	Name      string
	Icon      string
	SortOrder int32
	Level     int32
	Children  []*Category
	CreatedAt time.Time
}

func (c *Category) IsRoot() bool {
	return c.ParentID == 0
}

type CategoryRepository interface {
	Create(ctx context.Context, c *Category) (int64, error)
	Update(ctx context.Context, c *Category) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Category, error)
	GetTree(ctx context.Context, parentID int64) ([]*Category, error)
	HasChildren(ctx context.Context, id int64) (bool, error)
}

type CategoryBiz struct {
	repo CategoryRepository
}

func NewCategoryBiz(repo CategoryRepository) *CategoryBiz {
	return &CategoryBiz{repo: repo}
}

func (b *CategoryBiz) CreateCategory(ctx context.Context, name string, parentID int64, icon string, sortOrder int32) (*Category, error) {
	level := int32(1)
	if parentID > 0 {
		parent, err := b.repo.GetByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("get parent category: %w", err)
		}
		level = parent.Level + 1
		if level > 3 {
			return nil, fmt.Errorf("category level exceeds maximum (3)")
		}
	}

	cat := &Category{
		ParentID:  parentID,
		Name:      name,
		Icon:      icon,
		SortOrder: sortOrder,
		Level:     level,
	}

	id, err := b.repo.Create(ctx, cat)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	cat.ID = id
	return cat, nil
}

func (b *CategoryBiz) UpdateCategory(ctx context.Context, id int64, name, icon string, sortOrder int32) (*Category, error) {
	cat, err := b.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}

	cat.Name = name
	cat.Icon = icon
	cat.SortOrder = sortOrder

	if err := b.repo.Update(ctx, cat); err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return cat, nil
}

func (b *CategoryBiz) DeleteCategory(ctx context.Context, id int64) error {
	hasChildren, err := b.repo.HasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("check children: %w", err)
	}
	if hasChildren {
		return ErrCategoryHasChildren
	}
	return b.repo.Delete(ctx, id)
}

func (b *CategoryBiz) GetCategoryTree(ctx context.Context) ([]*Category, error) {
	roots, err := b.repo.GetTree(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("get category tree: %w", err)
	}

	for _, root := range roots {
		if err := b.populateChildren(ctx, root); err != nil {
			return nil, err
		}
	}
	return roots, nil
}

func (b *CategoryBiz) populateChildren(ctx context.Context, parent *Category) error {
	children, err := b.repo.GetTree(ctx, parent.ID)
	if err != nil {
		return fmt.Errorf("get children of category %d: %w", parent.ID, err)
	}
	if len(children) > 0 {
		parent.Children = children
		for _, child := range children {
			if err := b.populateChildren(ctx, child); err != nil {
				return err
			}
		}
	}
	return nil
}
