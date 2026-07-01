package biz

import (
	"context"
	"fmt"
	"time"
)

type CartItem struct {
	ID        int64
	UserID    int64
	ProductID int64
	SKUID     int64
	SPUID     int64
	ShopID    int64
	Title     string
	Image     string
	Attrs     map[string]string
	Price     float64
	Quantity  int32
	Selected  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (i *CartItem) SubTotal() float64 {
	return i.Price * float64(i.Quantity)
}

type CartRepository interface {
	AddItem(ctx context.Context, item *CartItem) error
	UpdateQuantity(ctx context.Context, userID, itemID int64, quantity int32) error
	RemoveItem(ctx context.Context, userID, itemID int64) error
	ListItems(ctx context.Context, userID int64) ([]*CartItem, error)
	ClearItems(ctx context.Context, userID int64, itemIDs []int64) error
}

type CartBiz struct {
	repo CartRepository
}

func NewCartBiz(repo CartRepository) *CartBiz {
	return &CartBiz{repo: repo}
}

func (b *CartBiz) AddItem(ctx context.Context, userID int64, item *CartItem) error {
	if item.Quantity <= 0 {
		return ErrCartQuantityInvalid
	}
	item.UserID = userID
	item.Selected = true
	if err := b.repo.AddItem(ctx, item); err != nil {
		return fmt.Errorf("add cart item: %w", err)
	}
	return nil
}

func (b *CartBiz) UpdateQuantity(ctx context.Context, userID, itemID int64, quantity int32) error {
	if quantity <= 0 {
		return ErrCartQuantityInvalid
	}
	if err := b.repo.UpdateQuantity(ctx, userID, itemID, quantity); err != nil {
		return fmt.Errorf("update cart quantity: %w", err)
	}
	return nil
}

func (b *CartBiz) RemoveItem(ctx context.Context, userID, itemID int64) error {
	if err := b.repo.RemoveItem(ctx, userID, itemID); err != nil {
		return fmt.Errorf("remove cart item: %w", err)
	}
	return nil
}

func (b *CartBiz) ListItems(ctx context.Context, userID int64) ([]*CartItem, error) {
	items, err := b.repo.ListItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list cart items: %w", err)
	}
	return items, nil
}

func (b *CartBiz) ClearItems(ctx context.Context, userID int64, itemIDs []int64) error {
	if len(itemIDs) == 0 {
		return nil
	}
	if err := b.repo.ClearItems(ctx, userID, itemIDs); err != nil {
		return fmt.Errorf("clear cart items: %w", err)
	}
	return nil
}
