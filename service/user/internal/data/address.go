package data

import (
	"context"
	"fmt"
	"time"

	"github.com/storm/myidea/service/user/internal/biz"
)

type GORMAddress struct {
	ID            int64     `gorm:"primaryKey;autoIncrement"`
	UserID        int64     `gorm:"index;not null"`
	ReceiverName  string    `gorm:"size:64;not null"`
	ReceiverPhone string    `gorm:"size:20;not null"`
	Province      string    `gorm:"size:64"`
	City          string    `gorm:"size:64"`
	District      string    `gorm:"size:64"`
	DetailAddress string    `gorm:"size:256;not null"`
	IsDefault     bool      `gorm:"default:false"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

func (GORMAddress) TableName() string {
	return "addresses"
}

type addressRepo struct {
	data *Data
}

func NewAddressRepo(data *Data) biz.AddressRepository {
	return &addressRepo{data: data}
}

func (r *addressRepo) Create(ctx context.Context, addr *biz.Address) (int64, error) {
	gormAddr := toGORMAddress(addr)
	if err := r.data.DB(ctx).Create(gormAddr).Error; err != nil {
		return 0, fmt.Errorf("insert address: %w", err)
	}
	return gormAddr.ID, nil
}

func (r *addressRepo) Update(ctx context.Context, addr *biz.Address) error {
	result := r.data.DB(ctx).Model(&GORMAddress{}).
		Where("id = ? AND user_id = ?", addr.ID, addr.UserID).
		Updates(map[string]any{
			"receiver_name":  addr.ReceiverName,
			"receiver_phone": addr.ReceiverPhone,
			"province":       addr.Province,
			"city":           addr.City,
			"district":       addr.District,
			"detail_address": addr.DetailAddress,
			"is_default":     addr.IsDefault,
		})
	if result.Error != nil {
		return fmt.Errorf("update address: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return biz.ErrAddressNotFound
	}
	return nil
}

func (r *addressRepo) Delete(ctx context.Context, id, userID int64) error {
	result := r.data.DB(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&GORMAddress{})
	if result.Error != nil {
		return fmt.Errorf("delete address: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return biz.ErrAddressNotFound
	}
	return nil
}

func (r *addressRepo) ListByUserID(ctx context.Context, userID int64) ([]*biz.Address, error) {
	var gormAddrs []GORMAddress
	if err := r.data.DB(ctx).Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&gormAddrs).Error; err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}

	addrs := make([]*biz.Address, len(gormAddrs))
	for i := range gormAddrs {
		addrs[i] = toDomainAddress(&gormAddrs[i])
	}
	return addrs, nil
}

func (r *addressRepo) SetDefault(ctx context.Context, id, userID int64) error {
	if err := r.data.DB(ctx).Model(&GORMAddress{}).
		Where("user_id = ?", userID).
		Update("is_default", false).Error; err != nil {
		return fmt.Errorf("clear defaults: %w", err)
	}

	if id > 0 {
		result := r.data.DB(ctx).Model(&GORMAddress{}).
			Where("id = ? AND user_id = ?", id, userID).
			Update("is_default", true)
		if result.Error != nil {
			return fmt.Errorf("set default: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return biz.ErrAddressNotFound
		}
	}
	return nil
}

func (r *addressRepo) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	if err := r.data.DB(ctx).Model(&GORMAddress{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count addresses: %w", err)
	}
	return count, nil
}

func toGORMAddress(a *biz.Address) *GORMAddress {
	return &GORMAddress{
		UserID:        a.UserID,
		ReceiverName:  a.ReceiverName,
		ReceiverPhone: a.ReceiverPhone,
		Province:      a.Province,
		City:          a.City,
		District:      a.District,
		DetailAddress: a.DetailAddress,
		IsDefault:     a.IsDefault,
	}
}

func toDomainAddress(g *GORMAddress) *biz.Address {
	return &biz.Address{
		ID:            g.ID,
		UserID:        g.UserID,
		ReceiverName:  g.ReceiverName,
		ReceiverPhone: g.ReceiverPhone,
		Province:      g.Province,
		City:          g.City,
		District:      g.District,
		DetailAddress: g.DetailAddress,
		IsDefault:     g.IsDefault,
		CreatedAt:     g.CreatedAt,
	}
}
