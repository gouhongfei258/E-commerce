package data

import (
	"context"
	"fmt"
	"time"

	"github.com/storm/myidea/service/user/internal/biz"
	"gorm.io/gorm"
)

type GORMUser struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Username  string    `gorm:"uniqueIndex;size:64;not null"`
	Password  string    `gorm:"size:255;not null"`
	Phone     string    `gorm:"size:20"`
	Email     string    `gorm:"size:128"`
	Role      string    `gorm:"size:16;default:user"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (GORMUser) TableName() string {
	return "users"
}

type userRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepository {
	return &userRepo{data: data}
}

func (r *userRepo) Create(ctx context.Context, user *biz.User) (int64, error) {
	gormUser := toGORMUser(user)
	if err := r.data.DB(ctx).Create(gormUser).Error; err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return gormUser.ID, nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	var gormUser GORMUser
	err := r.data.DB(ctx).Where("username = ?", username).First(&gormUser).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return toDomainUser(&gormUser), nil
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*biz.User, error) {
	var gormUser GORMUser
	err := r.data.DB(ctx).First(&gormUser, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return toDomainUser(&gormUser), nil
}

func (r *userRepo) List(ctx context.Context, keyword string, page, pageSize int32) ([]*biz.User, int32, error) {
	db := r.data.DB(ctx).Model(&GORMUser{})
	if keyword != "" {
		db = db.Where("username LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var pos []GORMUser
	if err := db.Order("id DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	users := make([]*biz.User, 0, len(pos))
	for i := range pos {
		users = append(users, toDomainUser(&pos[i]))
	}
	return users, int32(total), nil
}

func toGORMUser(u *biz.User) *GORMUser {
	return &GORMUser{
		Username: u.Username,
		Password: u.Password,
		Phone:    u.Phone,
		Email:    u.Email,
		Role:     u.Role,
	}
}

func toDomainUser(g *GORMUser) *biz.User {
	return &biz.User{
		ID:        g.ID,
		Username:  g.Username,
		Password:  g.Password,
		Phone:     g.Phone,
		Email:     g.Email,
		Role:      g.Role,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}
