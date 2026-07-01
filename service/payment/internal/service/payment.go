package service

import (
	"context"

	pb "github.com/storm/myidea/api/payment/v1"
	"github.com/storm/myidea/service/payment/internal/biz"
)

type PaymentService struct {
	pb.UnimplementedPaymentServiceServer
	biz *biz.PaymentBiz
}

func NewPaymentService(biz *biz.PaymentBiz) *PaymentService {
	return &PaymentService{biz: biz}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	payment, err := s.biz.CreatePayment(ctx, req.UserId, req.OrderNo, req.TotalAmount, biz.PaymentMethod(req.Method))
	if err != nil {
		return nil, err
	}
	return &pb.CreatePaymentResponse{Payment: paymentToProto(payment)}, nil
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	payment, err := s.biz.ProcessPayment(ctx, req.PaymentId, biz.PaymentMethod(req.Method))
	if err != nil {
		return nil, err
	}
	return &pb.ProcessPaymentResponse{Payment: paymentToProto(payment)}, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	payment, err := s.biz.GetPayment(ctx, req.PaymentId)
	if err != nil {
		return nil, err
	}
	return &pb.GetPaymentResponse{Payment: paymentToProto(payment)}, nil
}

func (s *PaymentService) GetPaymentByOrder(ctx context.Context, req *pb.GetPaymentByOrderRequest) (*pb.GetPaymentByOrderResponse, error) {
	payment, err := s.biz.GetPaymentByOrder(ctx, req.OrderNo)
	if err != nil {
		return nil, err
	}
	return &pb.GetPaymentByOrderResponse{Payment: paymentToProto(payment)}, nil
}

func (s *PaymentService) NotifyPayment(ctx context.Context, req *pb.NotifyPaymentRequest) (*pb.NotifyPaymentResponse, error) {
	payment, err := s.biz.NotifyPayment(ctx, req.PaymentId, biz.PaymentStatus(req.Status), req.ProviderTradeNo, req.ProviderRawResponse)
	if err != nil {
		return nil, err
	}
	return &pb.NotifyPaymentResponse{Payment: paymentToProto(payment)}, nil
}

func (s *PaymentService) AdminListPayments(ctx context.Context, req *pb.AdminListPaymentsRequest) (*pb.AdminListPaymentsResponse, error) {
	payments, total, err := s.biz.AdminListPayments(ctx, biz.PaymentStatus(req.Status), req.OrderNo, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	protos := make([]*pb.PaymentProto, len(payments))
	for i, p := range payments {
		protos[i] = paymentToProto(p)
	}
	return &pb.AdminListPaymentsResponse{Payments: protos, Total: total}, nil
}

func paymentToProto(p *biz.Payment) *pb.PaymentProto {
	if p == nil {
		return nil
	}

	var createdAt, updatedAt string
	if !p.CreatedAt.IsZero() {
		createdAt = p.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !p.UpdatedAt.IsZero() {
		updatedAt = p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return &pb.PaymentProto{
		Id:                  p.ID,
		UserId:              p.UserID,
		OrderNo:             p.OrderNo,
		TotalAmount:         p.TotalAmount,
		PaidAmount:          p.PaidAmount,
		Status:              pb.PaymentStatus(p.Status),
		Method:              pb.PaymentMethod(p.Method),
		ProviderTradeNo:     p.ProviderTradeNo,
		ProviderRawResponse: p.ProviderRawResponse,
		FailReason:          p.FailReason,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}
}
