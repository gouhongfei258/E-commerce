package biz

import (
	"context"
	"fmt"
	"time"
)

type Address struct {
	ID            int64
	UserID        int64
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	District      string
	DetailAddress string
	IsDefault     bool
	CreatedAt     time.Time
}

type AddressRepository interface {
	Create(ctx context.Context, addr *Address) (int64, error)
	Update(ctx context.Context, addr *Address) error
	Delete(ctx context.Context, id, userID int64) error
	ListByUserID(ctx context.Context, userID int64) ([]*Address, error)
	SetDefault(ctx context.Context, id, userID int64) error
	CountByUserID(ctx context.Context, userID int64) (int64, error)
}

type AddressBiz struct {
	repo AddressRepository
}

func NewAddressBiz(repo AddressRepository) *AddressBiz {
	return &AddressBiz{repo: repo}
}

func (b *AddressBiz) Create(ctx context.Context, addr *Address) (int64, error) {
	count, err := b.repo.CountByUserID(ctx, addr.UserID)
	if err != nil {
		return 0, fmt.Errorf("count user addresses: %w", err)
	}
	if count >= 10 {
		return 0, ErrAddressLimit
	}

	if addr.IsDefault {
		if err := b.repo.SetDefault(ctx, 0, addr.UserID); err != nil {
			return 0, fmt.Errorf("unset existing default: %w", err)
		}
	}

	id, err := b.repo.Create(ctx, addr)
	if err != nil {
		return 0, fmt.Errorf("create address: %w", err)
	}
	return id, nil
}

func (b *AddressBiz) Update(ctx context.Context, addr *Address) error {
	if err := b.repo.Update(ctx, addr); err != nil {
		return fmt.Errorf("update address: %w", err)
	}
	return nil
}

func (b *AddressBiz) Delete(ctx context.Context, id, userID int64) error {
	if err := b.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete address: %w", err)
	}
	return nil
}

func (b *AddressBiz) ListByUserID(ctx context.Context, userID int64) ([]*Address, error) {
	addresses, err := b.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	return addresses, nil
}

func (b *AddressBiz) SetDefault(ctx context.Context, id, userID int64) error {
	if err := b.repo.SetDefault(ctx, 0, userID); err != nil {
		return fmt.Errorf("unset existing default: %w", err)
	}
	if err := b.repo.SetDefault(ctx, id, userID); err != nil {
		return fmt.Errorf("set default address: %w", err)
	}
	return nil
}
