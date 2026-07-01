package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/storm/myidea/service/cart/internal/biz"
)

type GORMCartItem struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"index:idx_user_sku,unique,priority:1;not null"`
	ProductID int64     `gorm:"not null"`
	SKUID     int64     `gorm:"index:idx_user_sku,unique,priority:2;not null"`
	SPUID     int64     `gorm:"not null"`
	ShopID    int64     `gorm:"default:0"`
	Title     string    `gorm:"size:256;not null"`
	Image     string    `gorm:"size:512"`
	Attrs     string    `gorm:"type:json"`
	Price     float64   `gorm:"type:decimal(10,2);not null"`
	Quantity  int32     `gorm:"not null;default:1"`
	Selected  bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (GORMCartItem) TableName() string { return "cart_items" }

type cartRepo struct {
	data *Data
}

func NewCartRepo(data *Data) biz.CartRepository {
	return &cartRepo{data: data}
}

func (r *cartRepo) AddItem(ctx context.Context, item *biz.CartItem) error {
	attrsJSON, err := marshalAttrs(item.Attrs)
	if err != nil {
		return fmt.Errorf("marshal attrs: %w", err)
	}

	po := &GORMCartItem{
		UserID:    item.UserID,
		ProductID: item.ProductID,
		SKUID:     item.SKUID,
		SPUID:     item.SPUID,
		ShopID:    item.ShopID,
		Title:     item.Title,
		Image:     item.Image,
		Attrs:     attrsJSON,
		Price:     item.Price,
		Quantity:  item.Quantity,
		Selected:  item.Selected,
	}

	err = r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var existing GORMCartItem
		result := tx.Where("user_id = ? AND sku_id = ?", item.UserID, item.SKUID).First(&existing)
		if result.Error == nil {
			existing.Quantity += item.Quantity
			existing.Price = item.Price
			existing.Image = item.Image
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			item.ID = existing.ID
			item.Quantity = existing.Quantity
			return nil
		}
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}

		if err := tx.Create(po).Error; err != nil {
			return err
		}
		item.ID = po.ID
		return nil
	})
	if err != nil {
		return fmt.Errorf("add cart item: %w", err)
	}
	return nil
}

func (r *cartRepo) UpdateQuantity(ctx context.Context, userID, itemID int64, quantity int32) error {
	res := r.data.DB(ctx).Model(&GORMCartItem{}).
		Where("id = ? AND user_id = ?", itemID, userID).
		Update("quantity", quantity)
	if res.Error != nil {
		return fmt.Errorf("update cart quantity: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return biz.ErrCartItemNotFound
	}
	return nil
}

func (r *cartRepo) RemoveItem(ctx context.Context, userID, itemID int64) error {
	res := r.data.DB(ctx).Where("id = ? AND user_id = ?", itemID, userID).
		Delete(&GORMCartItem{})
	if res.Error != nil {
		return fmt.Errorf("remove cart item: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return biz.ErrCartItemNotFound
	}
	return nil
}

func (r *cartRepo) ListItems(ctx context.Context, userID int64) ([]*biz.CartItem, error) {
	var pos []GORMCartItem
	if err := r.data.DB(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("list cart items: %w", err)
	}

	items := make([]*biz.CartItem, 0, len(pos))
	for i := range pos {
		item, err := r.toDomain(&pos[i])
		if err != nil {
			return nil, fmt.Errorf("convert cart item %d: %w", pos[i].ID, err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *cartRepo) ClearItems(ctx context.Context, userID int64, itemIDs []int64) error {
	res := r.data.DB(ctx).Where("user_id = ? AND id IN ?", userID, itemIDs).
		Delete(&GORMCartItem{})
	if res.Error != nil {
		return fmt.Errorf("clear cart items: %w", res.Error)
	}
	return nil
}

func (r *cartRepo) toDomain(po *GORMCartItem) (*biz.CartItem, error) {
	attrs, err := unmarshalAttrs(po.Attrs)
	if err != nil {
		return nil, err
	}
	return &biz.CartItem{
		ID:        po.ID,
		UserID:    po.UserID,
		ProductID: po.ProductID,
		SKUID:     po.SKUID,
		SPUID:     po.SPUID,
		ShopID:    po.ShopID,
		Title:     po.Title,
		Image:     po.Image,
		Attrs:     attrs,
		Price:     po.Price,
		Quantity:  po.Quantity,
		Selected:  po.Selected,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}, nil
}

func marshalAttrs(attrs map[string]string) (string, error) {
	if len(attrs) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(attrs)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalAttrs(s string) (map[string]string, error) {
	if s == "" || s == "{}" {
		return nil, nil
	}
	var attrs map[string]string
	if err := json.Unmarshal([]byte(s), &attrs); err != nil {
		return nil, err
	}
	return attrs, nil
}
