package service

import (
	"context"

	pb "github.com/storm/myidea/api/order/v1"
	"github.com/storm/myidea/service/order/internal/biz"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer
	biz *biz.OrderBiz
}

func NewOrderService(biz *biz.OrderBiz) *OrderService {
	return &OrderService{biz: biz}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	items := make([]biz.OrderItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = biz.OrderItem{
			SKUID:       it.SkuId,
			ProductID:   it.ProductId,
			ProductName: it.ProductName,
			Image:       it.Image,
			Price:       it.Price,
			Quantity:    it.Quantity,
		}
	}

	addr := biz.ShippingAddress{
		ReceiverName:  req.Address.ReceiverName,
		ReceiverPhone: req.Address.ReceiverPhone,
		Province:      req.Address.Province,
		City:          req.Address.City,
		District:      req.Address.District,
		DetailAddress: req.Address.DetailAddress,
	}

	order, err := s.biz.CreateOrder(ctx, req.UserId, items, addr, req.Remark)
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrderResponse{Order: orderToProto(order)}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	order, err := s.biz.GetOrder(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.GetOrderResponse{Order: orderToProto(order)}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, total, err := s.biz.ListOrders(ctx, req.UserId, biz.OrderStatus(req.Status), req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	protos := make([]*pb.OrderProto, len(orders))
	for i, o := range orders {
		protos[i] = orderToProto(o)
	}
	return &pb.ListOrdersResponse{Orders: protos, Total: total}, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	err := s.biz.UpdateOrderStatus(ctx, req.Id, biz.OrderStatus(req.Status))
	if err != nil {
		return nil, err
	}
	return &pb.UpdateOrderStatusResponse{}, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	err := s.biz.CancelOrder(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.CancelOrderResponse{}, nil
}

func (s *OrderService) AdminListOrders(ctx context.Context, req *pb.AdminListOrdersRequest) (*pb.AdminListOrdersResponse, error) {
	orders, total, err := s.biz.AdminListOrders(ctx, biz.OrderStatus(req.Status), req.Keyword, req.DateFrom, req.DateTo, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	protos := make([]*pb.OrderProto, len(orders))
	for i, o := range orders {
		protos[i] = orderToProto(o)
	}
	return &pb.AdminListOrdersResponse{Orders: protos, Total: total}, nil
}

func orderToProto(o *biz.Order) *pb.OrderProto {
	items := make([]*pb.OrderItemProto, len(o.Items))
	for i, it := range o.Items {
		items[i] = &pb.OrderItemProto{
			SkuId:       it.SKUID,
			ProductId:   it.ProductID,
			ProductName: it.ProductName,
			Image:       it.Image,
			Price:       it.Price,
			Quantity:    it.Quantity,
		}
	}

	var createdAt, updatedAt string
	if !o.CreatedAt.IsZero() {
		createdAt = o.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !o.UpdatedAt.IsZero() {
		updatedAt = o.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return &pb.OrderProto{
		Id:      o.ID,
		UserId:  o.UserID,
		OrderNo: o.OrderNo,
		Status:  int32(o.Status),
		TotalAmount: o.TotalAmount,
		PaidAmount:  o.PaidAmount,
		PaymentMethod: o.PaymentMethod,
		Address: &pb.ShippingAddressProto{
			ReceiverName:  o.Address.ReceiverName,
			ReceiverPhone: o.Address.ReceiverPhone,
			Province:      o.Address.Province,
			City:          o.Address.City,
			District:      o.Address.District,
			DetailAddress: o.Address.DetailAddress,
		},
		Items:     items,
		Remark:    o.Remark,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
