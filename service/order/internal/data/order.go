package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/storm/myidea/service/order/internal/biz"
)

type GORMOrder struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	UserID        int64           `gorm:"index;not null"`
	OrderNo       string          `gorm:"uniqueIndex;size:32;not null"`
	Status        int32           `gorm:"default:0;not null"`
	TotalAmount   float64         `gorm:"type:decimal(12,2);not null"`
	PaidAmount    float64         `gorm:"type:decimal(12,2);default:0"`
	PaymentMethod string          `gorm:"size:32"`
	ReceiverName  string          `gorm:"size:64;not null"`
	ReceiverPhone string          `gorm:"size:20;not null"`
	Province      string          `gorm:"size:32"`
	City          string          `gorm:"size:32"`
	District      string          `gorm:"size:32"`
	DetailAddress string          `gorm:"size:256"`
	Remark        string          `gorm:"size:512"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`
	Items         []GORMOrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
}

func (GORMOrder) TableName() string { return "orders" }

type GORMOrderItem struct {
	ID          int64   `gorm:"primaryKey;autoIncrement"`
	OrderID     int64   `gorm:"index;not null"`
	SKUID       int64   `gorm:"default:0;index"`
	ProductID   int64   `gorm:"not null"`
	ProductName string  `gorm:"size:128;not null"`
	Image       string  `gorm:"size:512"`
	Price       float64 `gorm:"type:decimal(10,2);not null"`
	Quantity    int32   `gorm:"not null"`
}

func (GORMOrderItem) TableName() string { return "order_items" }

type orderRepo struct {
	data *Data
}

func NewOrderRepo(data *Data) biz.OrderRepository {
	return &orderRepo{data: data}
}

func (r *orderRepo) Save(ctx context.Context, order *biz.Order) (int64, error) {
	po := r.toGORM(order)
	err := r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		order.ID = po.ID
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}
	return po.ID, nil
}

func (r *orderRepo) GetByID(ctx context.Context, id int64) (*biz.Order, error) {
	var po GORMOrder
	err := r.data.DB(ctx).Preload("Items").First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	return r.toDomain(&po), nil
}

func (r *orderRepo) GetByIDAndUser(ctx context.Context, id, userID int64) (*biz.Order, error) {
	var po GORMOrder
	err := r.data.DB(ctx).Preload("Items").
		Where("id = ? AND user_id = ?", id, userID).
		First(&po).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id and user: %w", err)
	}
	return r.toDomain(&po), nil
}

func (r *orderRepo) List(ctx context.Context, userID int64, status biz.OrderStatus, page, pageSize int32) ([]*biz.Order, int32, error) {
	db := r.data.DB(ctx).Model(&GORMOrder{}).Where("user_id = ?", userID)
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var pos []GORMOrder
	if err := db.Preload("Items").Order("id DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}

	orders := make([]*biz.Order, 0, len(pos))
	for i := range pos {
		orders = append(orders, r.toDomain(&pos[i]))
	}
	return orders, int32(total), nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id int64, status biz.OrderStatus) error {
	res := r.data.DB(ctx).Model(&GORMOrder{}).Where("id = ?", id).
		Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("update order status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return biz.ErrOrderNotFound
	}
	return nil
}

func (r *orderRepo) AdminList(ctx context.Context, status biz.OrderStatus, keyword, dateFrom, dateTo string, page, pageSize int32) ([]*biz.Order, int32, error) {
	db := r.data.DB(ctx).Model(&GORMOrder{})
	if status > 0 {
		db = db.Where("status = ?", status)
	}
	if keyword != "" {
		db = db.Where("order_no LIKE ? OR receiver_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if dateFrom != "" {
		db = db.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		db = db.Where("created_at <= ?", dateTo+" 23:59:59")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var pos []GORMOrder
	if err := db.Preload("Items").Order("id DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("admin list orders: %w", err)
	}

	orders := make([]*biz.Order, 0, len(pos))
	for i := range pos {
		orders = append(orders, r.toDomain(&pos[i]))
	}
	return orders, int32(total), nil
}

func (r *orderRepo) toGORM(domain *biz.Order) *GORMOrder {
	items := make([]GORMOrderItem, len(domain.Items))
	for i, item := range domain.Items {
		items[i] = GORMOrderItem{
			SKUID:       item.SKUID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Image:       item.Image,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}
	}
	return &GORMOrder{
		ID:            domain.ID,
		UserID:        domain.UserID,
		OrderNo:       domain.OrderNo,
		Status:        int32(domain.Status),
		TotalAmount:   domain.TotalAmount,
		PaidAmount:    domain.PaidAmount,
		PaymentMethod: domain.PaymentMethod,
		ReceiverName:  domain.Address.ReceiverName,
		ReceiverPhone: domain.Address.ReceiverPhone,
		Province:      domain.Address.Province,
		City:          domain.Address.City,
		District:      domain.Address.District,
		DetailAddress: domain.Address.DetailAddress,
		Remark:        domain.Remark,
		Items:         items,
	}
}

func (r *orderRepo) toDomain(po *GORMOrder) *biz.Order {
	items := make([]biz.OrderItem, len(po.Items))
	for i, item := range po.Items {
		items[i] = biz.OrderItem{
			SKUID:       item.SKUID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Image:       item.Image,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}
	}
	return &biz.Order{
		ID:            po.ID,
		UserID:        po.UserID,
		OrderNo:       po.OrderNo,
		Status:        biz.OrderStatus(po.Status),
		TotalAmount:   po.TotalAmount,
		PaidAmount:    po.PaidAmount,
		PaymentMethod: po.PaymentMethod,
		Address: biz.ShippingAddress{
			ReceiverName:  po.ReceiverName,
			ReceiverPhone: po.ReceiverPhone,
			Province:      po.Province,
			City:          po.City,
			District:      po.District,
			DetailAddress: po.DetailAddress,
		},
		Remark:    po.Remark,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
		Items:     items,
	}
}
