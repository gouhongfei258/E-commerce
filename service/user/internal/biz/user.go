package biz

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64
	Username  string
	Password  string
	Phone     string
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(bytes), nil
}

func CheckPassword(plain, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return ErrUserPasswordWrong
	}
	return nil
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (int64, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	List(ctx context.Context, keyword string, page, pageSize int32) ([]*User, int32, error)
}

type UserBiz struct {
	repo UserRepository
}

func NewUserBiz(repo UserRepository) *UserBiz {
	return &UserBiz{repo: repo}
}

func (b *UserBiz) Register(ctx context.Context, username, password, phone, email, role string) (*User, error) {
	existing, err := b.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserDuplicate
	}

	if role == "" {
		role = "user"
	}

	hashed, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	user := &User{
		Username: username,
		Password: hashed,
		Phone:    phone,
		Email:    email,
		Role:     role,
	}

	id, err := b.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.ID = id
	return user, nil
}

func (b *UserBiz) Login(ctx context.Context, username, password string) (*User, error) {
	user, err := b.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := CheckPassword(password, user.Password); err != nil {
		return nil, err
	}
	return user, nil
}

func (b *UserBiz) GetUser(ctx context.Context, id int64) (*User, error) {
	user, err := b.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (b *UserBiz) AdminListUsers(ctx context.Context, keyword string, page, pageSize int32) ([]*User, int32, error) {
	return b.repo.List(ctx, keyword, page, pageSize)
}
