package biz

import (
	"context"
	"fmt"
	"time"
)

type PaymentStatus int32

const (
	PaymentStatusPending   PaymentStatus = 1
	PaymentStatusSuccess   PaymentStatus = 2
	PaymentStatusFailed    PaymentStatus = 3
	PaymentStatusRefunding PaymentStatus = 4
	PaymentStatusRefunded  PaymentStatus = 5
)

func (s PaymentStatus) String() string {
	switch s {
	case PaymentStatusPending:
		return "pending"
	case PaymentStatusSuccess:
		return "success"
	case PaymentStatusFailed:
		return "failed"
	case PaymentStatusRefunding:
		return "refunding"
	case PaymentStatusRefunded:
		return "refunded"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

func (s PaymentStatus) CanTransitionTo(target PaymentStatus) bool {
	transitions := map[PaymentStatus][]PaymentStatus{
		PaymentStatusPending:   {PaymentStatusSuccess, PaymentStatusFailed},
		PaymentStatusSuccess:   {PaymentStatusRefunding},
		PaymentStatusFailed:    {},
		PaymentStatusRefunding: {PaymentStatusRefunded},
		PaymentStatusRefunded:  {},
	}
	next, ok := transitions[s]
	if !ok {
		return false
	}
	for _, n := range next {
		if n == target {
			return true
		}
	}
	return false
}

type PaymentMethod int32

const (
	PaymentMethodMock   PaymentMethod = 1
	PaymentMethodAlipay PaymentMethod = 2
	PaymentMethodWechat PaymentMethod = 3
)

func (m PaymentMethod) String() string {
	switch m {
	case PaymentMethodMock:
		return "mock"
	case PaymentMethodAlipay:
		return "alipay"
	case PaymentMethodWechat:
		return "wechat_pay"
	default:
		return fmt.Sprintf("unknown(%d)", m)
	}
}

type Payment struct {
	ID                  int64
	UserID              int64
	OrderNo             string
	TotalAmount         float64
	PaidAmount          float64
	Status              PaymentStatus
	Method              PaymentMethod
	ProviderTradeNo     string
	ProviderRawResponse string
	FailReason          string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type PaymentProvider interface {
	Pay(ctx context.Context, req *PaymentProviderRequest) (*PaymentProviderResponse, error)
	VerifySignature(ctx context.Context, rawResponse string) (bool, error)
	Name() string
}

type PaymentProviderRequest struct {
	OrderNo     string
	TotalAmount float64
	Description string
}

type PaymentProviderResponse struct {
	ProviderTradeNo string
	RawResponse     string
	Success         bool
	FailReason      string
}

type PaymentProviderFactory interface {
	GetProvider(method PaymentMethod) (PaymentProvider, error)
	Register(method PaymentMethod, provider PaymentProvider)
}

type PaymentRepository interface {
	Save(ctx context.Context, payment *Payment) (int64, error)
	GetByID(ctx context.Context, id int64) (*Payment, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*Payment, error)
	UpdateStatus(ctx context.Context, id int64, status PaymentStatus, paidAmount float64, providerTradeNo, providerRawResponse, failReason string) error
	AdminList(ctx context.Context, status PaymentStatus, orderNo string, page, pageSize int32) ([]*Payment, int32, error)
}

type PaymentBiz struct {
	repo            PaymentRepository
	providerFactory PaymentProviderFactory
}

func NewPaymentBiz(repo PaymentRepository, providerFactory PaymentProviderFactory) *PaymentBiz {
	return &PaymentBiz{
		repo:            repo,
		providerFactory: providerFactory,
	}
}

func (b *PaymentBiz) CreatePayment(ctx context.Context, userID int64, orderNo string, totalAmount float64, method PaymentMethod) (*Payment, error) {
	if method == PaymentMethodMock || method == PaymentMethodAlipay || method == PaymentMethodWechat {
	} else {
		return nil, ErrPaymentInvalidMethod
	}

	payment := &Payment{
		UserID:      userID,
		OrderNo:     orderNo,
		TotalAmount: totalAmount,
		Status:      PaymentStatusPending,
		Method:      method,
	}

	id, err := b.repo.Save(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("save payment: %w", err)
	}
	payment.ID = id
	return payment, nil
}

func (b *PaymentBiz) ProcessPayment(ctx context.Context, paymentID int64, method PaymentMethod) (*Payment, error) {
	payment, err := b.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}

	if payment.Status != PaymentStatusPending {
		return nil, ErrPaymentAlreadyProcessed
	}

	pm := method
	if pm == 0 {
		pm = payment.Method
	}

	provider, err := b.providerFactory.GetProvider(pm)
	if err != nil {
		return nil, ErrPaymentInvalidMethod
	}

	providerResp, err := provider.Pay(ctx, &PaymentProviderRequest{
		OrderNo:     payment.OrderNo,
		TotalAmount: payment.TotalAmount,
		Description: fmt.Sprintf("Payment for order %s", payment.OrderNo),
	})
	if err != nil {
		b.repo.UpdateStatus(ctx, payment.ID, PaymentStatusFailed, 0, "", providerResp.RawResponse, fmt.Sprintf("provider error: %v", err))
		return nil, fmt.Errorf("payment provider: %w", err)
	}

	if providerResp.Success {
		err = b.repo.UpdateStatus(ctx, payment.ID, PaymentStatusSuccess, payment.TotalAmount, providerResp.ProviderTradeNo, providerResp.RawResponse, "")
	} else {
		err = b.repo.UpdateStatus(ctx, payment.ID, PaymentStatusFailed, 0, providerResp.ProviderTradeNo, providerResp.RawResponse, providerResp.FailReason)
	}
	if err != nil {
		return nil, fmt.Errorf("update payment status: %w", err)
	}

	return b.repo.GetByID(ctx, payment.ID)
}

func (b *PaymentBiz) GetPayment(ctx context.Context, id int64) (*Payment, error) {
	return b.repo.GetByID(ctx, id)
}

func (b *PaymentBiz) GetPaymentByOrder(ctx context.Context, orderNo string) (*Payment, error) {
	return b.repo.GetByOrderNo(ctx, orderNo)
}

func (b *PaymentBiz) NotifyPayment(ctx context.Context, paymentID int64, status PaymentStatus, providerTradeNo, rawResponse string) (*Payment, error) {
	payment, err := b.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("get payment for notify: %w", err)
	}

	if payment.Status != PaymentStatusPending {
		return nil, ErrPaymentAlreadyProcessed
	}

	provider, err := b.providerFactory.GetProvider(payment.Method)
	if err == nil {
		valid, sigErr := provider.VerifySignature(ctx, rawResponse)
		if sigErr == nil && !valid {
			return nil, ErrPaymentProviderFail
		}
	}

	var paidAmount float64
	if status == PaymentStatusSuccess {
		paidAmount = payment.TotalAmount
	}

	err = b.repo.UpdateStatus(ctx, payment.ID, status, paidAmount, providerTradeNo, rawResponse, "")
	if err != nil {
		return nil, fmt.Errorf("update payment after notify: %w", err)
	}

	return b.repo.GetByID(ctx, payment.ID)
}

type DefaultPaymentProviderFactory struct {
	providers map[PaymentMethod]PaymentProvider
}

func NewDefaultPaymentProviderFactory() *DefaultPaymentProviderFactory {
	return &DefaultPaymentProviderFactory{
		providers: make(map[PaymentMethod]PaymentProvider),
	}
}

func (f *DefaultPaymentProviderFactory) GetProvider(method PaymentMethod) (PaymentProvider, error) {
	p, ok := f.providers[method]
	if !ok {
		return nil, fmt.Errorf("no provider registered for method %s", method)
	}
	return p, nil
}

func (f *DefaultPaymentProviderFactory) Register(method PaymentMethod, provider PaymentProvider) {
	f.providers[method] = provider
}

func (b *PaymentBiz) AdminListPayments(ctx context.Context, status PaymentStatus, orderNo string, page, pageSize int32) ([]*Payment, int32, error) {
	return b.repo.AdminList(ctx, status, orderNo, page, pageSize)
}
