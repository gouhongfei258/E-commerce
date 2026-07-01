package biz

import (
	"context"
	"fmt"
	"time"
)

type OrderStatus int32

const (
	OrderStatusPending   OrderStatus = 0
	OrderStatusPaid      OrderStatus = 1
	OrderStatusShipped   OrderStatus = 2
	OrderStatusDelivered OrderStatus = 3
	OrderStatusCancelled OrderStatus = 4
	OrderStatusRefunding OrderStatus = 5
	OrderStatusRefunded  OrderStatus = 6
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "pending"
	case OrderStatusPaid:
		return "paid"
	case OrderStatusShipped:
		return "shipped"
	case OrderStatusDelivered:
		return "delivered"
	case OrderStatusCancelled:
		return "cancelled"
	case OrderStatusRefunding:
		return "refunding"
	case OrderStatusRefunded:
		return "refunded"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:   {OrderStatusPaid, OrderStatusCancelled},
		OrderStatusPaid:      {OrderStatusShipped, OrderStatusRefunding},
		OrderStatusShipped:   {OrderStatusDelivered},
		OrderStatusDelivered: {OrderStatusRefunding},
		OrderStatusRefunding: {OrderStatusRefunded},
		OrderStatusCancelled: {},
		OrderStatusRefunded:  {},
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

type ShippingAddress struct {
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	District      string
	DetailAddress string
}

type OrderItem struct {
	SKUID       int64
	ProductID   int64
	ProductName string
	Image       string
	Price       float64
	Quantity    int32
}

func (i *OrderItem) SubTotal() float64 {
	return i.Price * float64(i.Quantity)
}

type Order struct {
	ID            int64
	UserID        int64
	OrderNo       string
	Status        OrderStatus
	TotalAmount   float64
	PaidAmount    float64
	PaymentMethod string
	Address       ShippingAddress
	Items         []OrderItem
	Remark        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (o *Order) CalculateTotal() float64 {
	var total float64
	for i := range o.Items {
		total += o.Items[i].SubTotal()
	}
	return total
}

type OrderRepository interface {
	Save(ctx context.Context, order *Order) (int64, error)
	GetByID(ctx context.Context, id int64) (*Order, error)
	GetByIDAndUser(ctx context.Context, id, userID int64) (*Order, error)
	List(ctx context.Context, userID int64, status OrderStatus, page, pageSize int32) ([]*Order, int32, error)
	UpdateStatus(ctx context.Context, id int64, status OrderStatus) error
	AdminList(ctx context.Context, status OrderStatus, keyword, dateFrom, dateTo string, page, pageSize int32) ([]*Order, int32, error)
}

type ProductStockClient interface {
	LockStock(ctx context.Context, skuID int64, quantity int32, orderNo string) error
	ConfirmDeductStock(ctx context.Context, skuID int64, quantity int32, orderNo string) error
	UnlockStock(ctx context.Context, skuID int64, quantity int32, orderNo string) error
}

type OrderBiz struct {
	repo               OrderRepository
	productStockClient ProductStockClient
}

func NewOrderBiz(repo OrderRepository, productStockClient ProductStockClient) *OrderBiz {
	return &OrderBiz{
		repo:               repo,
		productStockClient: productStockClient,
	}
}

func (b *OrderBiz) CreateOrder(ctx context.Context, userID int64, items []OrderItem, addr ShippingAddress, remark string) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrOrderItemEmpty
	}
	if addr.ReceiverName == "" || addr.ReceiverPhone == "" {
		return nil, ErrOrderAddressRequired
	}

	order := &Order{
		UserID:  userID,
		OrderNo: generateOrderNo(userID),
		Status:  OrderStatusPending,
		Address: addr,
		Items:   items,
		Remark:  remark,
	}
	order.TotalAmount = order.CalculateTotal()

	for _, item := range items {
		if item.SKUID > 0 {
			if err := b.productStockClient.LockStock(ctx, item.SKUID, item.Quantity, order.OrderNo); err != nil {
				b.unlockAll(ctx, items, order.OrderNo)
				return nil, err
			}
		}
	}

	id, err := b.repo.Save(ctx, order)
	if err != nil {
		b.unlockAll(ctx, items, order.OrderNo)
		return nil, fmt.Errorf("save order: %w", err)
	}
	order.ID = id
	return order, nil
}

func (b *OrderBiz) unlockAll(ctx context.Context, items []OrderItem, orderNo string) {
	for _, item := range items {
		if item.SKUID > 0 {
			_ = b.productStockClient.UnlockStock(ctx, item.SKUID, item.Quantity, orderNo)
		}
	}
}

func (b *OrderBiz) GetOrder(ctx context.Context, id, userID int64) (*Order, error) {
	o, err := b.repo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return o, nil
}

func (b *OrderBiz) ListOrders(ctx context.Context, userID int64, status OrderStatus, page, pageSize int32) ([]*Order, int32, error) {
	return b.repo.List(ctx, userID, status, page, pageSize)
}

func (b *OrderBiz) UpdateOrderStatus(ctx context.Context, id int64, status OrderStatus) error {
	o, err := b.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get order for status update: %w", err)
	}
	if !o.Status.CanTransitionTo(status) {
		return ErrOrderStatusInvalid
	}

	if status == OrderStatusPaid {
		for _, item := range o.Items {
			if item.SKUID > 0 {
				if err := b.productStockClient.ConfirmDeductStock(ctx, item.SKUID, item.Quantity, o.OrderNo); err != nil {
					return fmt.Errorf("confirm deduct stock: %w", err)
				}
			}
		}
	}

	return b.repo.UpdateStatus(ctx, id, status)
}

func (b *OrderBiz) CancelOrder(ctx context.Context, id, userID int64) error {
	o, err := b.repo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("get order for cancel: %w", err)
	}
	if o.Status != OrderStatusPending {
		return ErrOrderCannotCancel
	}

	for _, item := range o.Items {
		if item.SKUID > 0 {
			_ = b.productStockClient.UnlockStock(ctx, item.SKUID, item.Quantity, o.OrderNo)
		}
	}

	return b.repo.UpdateStatus(ctx, id, OrderStatusCancelled)
}

func generateOrderNo(userID int64) string {
	now := time.Now().Format("20060102")
	return fmt.Sprintf("%s%08d%06d", now, userID, time.Now().UnixNano()%1000000)
}

func (b *OrderBiz) AdminListOrders(ctx context.Context, status OrderStatus, keyword, dateFrom, dateTo string, page, pageSize int32) ([]*Order, int32, error) {
	return b.repo.AdminList(ctx, status, keyword, dateFrom, dateTo, page, pageSize)
}
