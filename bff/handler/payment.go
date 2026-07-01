package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/storm/myidea/api/payment/v1"
)

// PaymentHandler handles HTTP requests for the payment domain.
type PaymentHandler struct {
	client pb.PaymentServiceClient
	conn   *grpc.ClientConn
}

// NewPaymentHandler creates a new handler and dials the gRPC connection.
func NewPaymentHandler(grpcAddr string, dialTimeout time.Duration) (*PaymentHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryClientInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	return &PaymentHandler{
		client: pb.NewPaymentServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close shuts down the gRPC connection.
func (h *PaymentHandler) Close() error {
	return h.conn.Close()
}

// PaymentServiceClient exposes the underlying gRPC client for cross-handler use (e.g. checkout).
func (h *PaymentHandler) PaymentServiceClient() pb.PaymentServiceClient {
	return h.client
}

// CreatePayment  POST /api/v1/payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		OrderNo     string  `json:"order_no" binding:"required"`
		TotalAmount float64 `json:"total_amount" binding:"required,gt=0"`
		Method      int32   `json:"method" binding:"omitempty,oneof=1 2 3"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	// Default to MOCK if not specified.
	if req.Method == 0 {
		req.Method = 1 // MOCK
	}

	ctx := injectUserID(c, userID)
	resp, err := h.client.CreatePayment(ctx, &pb.CreatePaymentRequest{
		UserId:      userID,
		OrderNo:     req.OrderNo,
		TotalAmount: req.TotalAmount,
		Method:      pb.PaymentMethod(req.Method),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Payment)
}

// ProcessPayment  POST /api/v1/payments/:id/process
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid payment id", nil)
		return
	}

	var req struct {
		Method int32 `json:"method" binding:"omitempty,oneof=1 2 3"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := injectUserID(c, userID)
	resp, err := h.client.ProcessPayment(ctx, &pb.ProcessPaymentRequest{
		PaymentId: id,
		Method:    pb.PaymentMethod(req.Method),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Payment)
}

// GetPayment  GET /api/v1/payments/:id
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid payment id", nil)
		return
	}

	ctx := injectUserID(c, userID)
	resp, err := h.client.GetPayment(ctx, &pb.GetPaymentRequest{
		PaymentId: id,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Payment)
}

// GetPaymentByOrder  GET /api/v1/payments/by-order/:orderNo
func (h *PaymentHandler) GetPaymentByOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orderNo := c.Param("orderNo")

	ctx := injectUserID(c, userID)
	resp, err := h.client.GetPaymentByOrder(ctx, &pb.GetPaymentByOrderRequest{
		OrderNo: orderNo,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Payment)
}

// NotifyPayment  POST /api/v1/payments/:id/notify
// This endpoint simulates a third-party payment provider callback.
// In production, each provider's webhook URL would be different and would include
// signature verification specific to that provider.
func (h *PaymentHandler) NotifyPayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid payment id", nil)
		return
	}

	var req struct {
		Status              int32  `json:"status" binding:"required,oneof=2 3"`
		ProviderTradeNo     string `json:"provider_trade_no"`
		ProviderRawResponse string `json:"provider_raw_response"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	// Notify endpoint does not require user auth — it simulates a third-party callback.
	ctx := metadata.NewOutgoingContext(c.Request.Context(), metadata.Pairs("x-user-id", "0"))
	resp, err := h.client.NotifyPayment(ctx, &pb.NotifyPaymentRequest{
		PaymentId:           id,
		Status:              pb.PaymentStatus(req.Status),
		ProviderTradeNo:     req.ProviderTradeNo,
		ProviderRawResponse: req.ProviderRawResponse,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Payment)
}

// ----------- gRPC error helpers (same as order.go) -----------

func paymentInjectUserID(c *gin.Context, userID int64) context.Context {
	ctx := c.Request.Context()
	md := metadata.Pairs("x-user-id", strconv.FormatInt(userID, 10))
	return metadata.NewOutgoingContext(ctx, md)
}

func paymentRespond(c *gin.Context, httpStatus, code int, msg string, data any) {
	c.JSON(httpStatus, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

func paymentRespondError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		paymentRespond(c, http.StatusInternalServerError, 500, "internal server error", nil)
		return
	}
	httpStatus := grpcCodeToHTTP(uint32(st.Code()))
	paymentRespond(c, httpStatus, int(st.Code()), st.Message(), nil)
}
