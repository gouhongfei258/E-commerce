package data

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"

	"github.com/storm/myidea/service/payment/internal/biz"
)

type GORMPayment struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement"`
	UserID              int64     `gorm:"index;not null"`
	OrderNo             string    `gorm:"uniqueIndex:uk_order_no;size:32;not null"`
	TotalAmount         float64   `gorm:"type:decimal(12,2);not null"`
	PaidAmount          float64   `gorm:"type:decimal(12,2);default:0"`
	Status              int32     `gorm:"default:1;not null"`
	Method              int32     `gorm:"default:1;not null"`
	ProviderTradeNo     string    `gorm:"size:128"`
	ProviderRawResponse string    `gorm:"type:text"`
	FailReason          string    `gorm:"size:512"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"`
}

func (GORMPayment) TableName() string { return "payments" }

type paymentRepo struct {
	data *Data
}

func NewPaymentRepo(data *Data) biz.PaymentRepository {
	return &paymentRepo{data: data}
}

func (r *paymentRepo) Save(ctx context.Context, payment *biz.Payment) (int64, error) {
	po := r.toGORM(payment)
	if err := r.data.DB(ctx).Create(po).Error; err != nil {
		return 0, fmt.Errorf("create payment: %w", err)
	}
	return po.ID, nil
}

func (r *paymentRepo) GetByID(ctx context.Context, id int64) (*biz.Payment, error) {
	var po GORMPayment
	err := r.data.DB(ctx).First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("get payment by id: %w", err)
	}
	return r.toDomain(&po), nil
}

func (r *paymentRepo) GetByOrderNo(ctx context.Context, orderNo string) (*biz.Payment, error) {
	var po GORMPayment
	err := r.data.DB(ctx).Where("order_no = ?", orderNo).First(&po).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("get payment by order no: %w", err)
	}
	return r.toDomain(&po), nil
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, id int64, status biz.PaymentStatus, paidAmount float64, providerTradeNo, providerRawResponse, failReason string) error {
	updates := map[string]any{
		"status":                int32(status),
		"paid_amount":           paidAmount,
		"provider_trade_no":     providerTradeNo,
		"provider_raw_response": providerRawResponse,
		"fail_reason":           failReason,
	}
	res := r.data.DB(ctx).Model(&GORMPayment{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update payment status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return biz.ErrPaymentNotFound
	}
	return nil
}

func (r *paymentRepo) AdminList(ctx context.Context, status biz.PaymentStatus, orderNo string, page, pageSize int32) ([]*biz.Payment, int32, error) {
	db := r.data.DB(ctx).Model(&GORMPayment{})
	if status > 0 {
		db = db.Where("status = ?", int32(status))
	}
	if orderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+orderNo+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count payments: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var pos []GORMPayment
	if err := db.Order("id DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("admin list payments: %w", err)
	}

	payments := make([]*biz.Payment, 0, len(pos))
	for i := range pos {
		payments = append(payments, r.toDomain(&pos[i]))
	}
	return payments, int32(total), nil
}

func (r *paymentRepo) toGORM(domain *biz.Payment) *GORMPayment {
	return &GORMPayment{
		UserID:      domain.UserID,
		OrderNo:     domain.OrderNo,
		TotalAmount: domain.TotalAmount,
		PaidAmount:  domain.PaidAmount,
		Status:      int32(domain.Status),
		Method:      int32(domain.Method),
	}
}

func (r *paymentRepo) toDomain(po *GORMPayment) *biz.Payment {
	return &biz.Payment{
		ID:                  po.ID,
		UserID:              po.UserID,
		OrderNo:             po.OrderNo,
		TotalAmount:         po.TotalAmount,
		PaidAmount:          po.PaidAmount,
		Status:              biz.PaymentStatus(po.Status),
		Method:              biz.PaymentMethod(po.Method),
		ProviderTradeNo:     po.ProviderTradeNo,
		ProviderRawResponse: po.ProviderRawResponse,
		FailReason:          po.FailReason,
		CreatedAt:           po.CreatedAt,
		UpdatedAt:           po.UpdatedAt,
	}
}

type MockPaymentProvider struct{}

func NewMockPaymentProvider() *MockPaymentProvider {
	return &MockPaymentProvider{}
}

func (p *MockPaymentProvider) Pay(ctx context.Context, req *biz.PaymentProviderRequest) (*biz.PaymentProviderResponse, error) {
	time.Sleep(time.Duration(50+rand.Intn(200)) * time.Millisecond)

	success := rand.Float64() < 0.8

	resp := &biz.PaymentProviderResponse{
		ProviderTradeNo: fmt.Sprintf("MOCK%d", time.Now().UnixNano()),
		RawResponse:     fmt.Sprintf(`{"mock":true,"order_no":"%s","amount":%.2f,"success":%t}`, req.OrderNo, req.TotalAmount, success),
		Success:         success,
	}

	if !success {
		resp.FailReason = "mock provider: simulated payment failure"
	}

	return resp, nil
}

func (p *MockPaymentProvider) VerifySignature(ctx context.Context, rawResponse string) (bool, error) {
	return true, nil
}

func (p *MockPaymentProvider) Name() string {
	return "mock"
}
