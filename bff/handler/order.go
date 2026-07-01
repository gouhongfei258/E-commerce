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

	pb "github.com/storm/myidea/api/order/v1"
)

// OrderHandler handles HTTP requests for the order domain.
// It translates HTTP ↔ gRPC and owns no business logic.
type OrderHandler struct {
	client pb.OrderServiceClient
	conn   *grpc.ClientConn
}

// NewOrderHandler creates a new handler and dials the gRPC connection.
func NewOrderHandler(grpcAddr string, dialTimeout time.Duration) (*OrderHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryClientInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	return &OrderHandler{
		client: pb.NewOrderServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close shuts down the gRPC connection.
func (h *OrderHandler) Close() error {
	return h.conn.Close()
}

// OrderServiceClient exposes the underlying gRPC client for cross-handler use (e.g. cart checkout).
func (h *OrderHandler) OrderServiceClient() pb.OrderServiceClient {
	return h.client
}

// CreateOrder  POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Items  []struct {
			SKUID       int64   `json:"sku_id"`
			ProductID   int64   `json:"product_id" binding:"required"`
			ProductName string  `json:"product_name" binding:"required"`
			Image       string  `json:"image"`
			Price       float64 `json:"price" binding:"required,gt=0"`
			Quantity    int32   `json:"quantity" binding:"required,gt=0"`
		} `json:"items" binding:"required,min=1"`
		Address struct {
			ReceiverName  string `json:"receiver_name" binding:"required"`
			ReceiverPhone string `json:"receiver_phone" binding:"required"`
			Province      string `json:"province"`
			City          string `json:"city"`
			District      string `json:"district"`
			DetailAddress string `json:"detail_address" binding:"required"`
		} `json:"address" binding:"required"`
		Remark string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	items := make([]*pb.OrderItemProto, len(req.Items))
	for i, it := range req.Items {
		items[i] = &pb.OrderItemProto{
			SkuId:       it.SKUID,
			ProductId:   it.ProductID,
			ProductName: it.ProductName,
			Image:       it.Image,
			Price:       it.Price,
			Quantity:    it.Quantity,
		}
	}

	ctx := injectUserID(c, userID)
	resp, err := h.client.CreateOrder(ctx, &pb.CreateOrderRequest{
		UserId: userID,
		Items:  items,
		Address: &pb.ShippingAddressProto{
			ReceiverName:  req.Address.ReceiverName,
			ReceiverPhone: req.Address.ReceiverPhone,
			Province:      req.Address.Province,
			City:          req.Address.City,
			District:      req.Address.District,
			DetailAddress: req.Address.DetailAddress,
		},
		Remark: req.Remark,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Order)
}

// GetOrder  GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid order id", nil)
		return
	}

	ctx := injectUserID(c, userID)
	resp, err := h.client.GetOrder(ctx, &pb.GetOrderRequest{
		Id:     id,
		UserId: userID,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", resp.Order)
}

// ListOrders  GET /api/v1/orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := c.GetInt64("user_id")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	ctx := injectUserID(c, userID)
	resp, err := h.client.ListOrders(ctx, &pb.ListOrdersRequest{
		UserId:   userID,
		Status:   int32(status),
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{
		"orders": resp.Orders,
		"total":  resp.Total,
	})
}

// UpdateOrderStatus  PUT /api/v1/orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid order id", nil)
		return
	}

	var req struct {
		Status int32 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := injectUserID(c, c.GetInt64("user_id"))
	_, err = h.client.UpdateOrderStatus(ctx, &pb.UpdateOrderStatusRequest{
		Id:     id,
		Status: req.Status,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", nil)
}

// CancelOrder  POST /api/v1/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid order id", nil)
		return
	}

	ctx := injectUserID(c, userID)
	_, err = h.client.CancelOrder(ctx, &pb.CancelOrderRequest{
		Id:     id,
		UserId: userID,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", nil)
}

// -------------------- Helpers --------------------

// unaryClientInterceptor attaches basic observability to every gRPC call.
// In production, use otelgrpc.UnaryClientInterceptor.
func unaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) error {
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// injectUserID propagates the caller identity to the downstream service
// via gRPC metadata.
func injectUserID(c *gin.Context, userID int64) context.Context {
	ctx := c.Request.Context()
	md := metadata.Pairs("x-user-id", strconv.FormatInt(userID, 10))
	return metadata.NewOutgoingContext(ctx, md)
}

// respond writes the standard JSON envelope.
func respond(c *gin.Context, httpStatus, code int, msg string, data any) {
	c.JSON(httpStatus, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

// respondError converts a gRPC status error into the standard JSON error envelope.
func respondError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		respond(c, http.StatusInternalServerError, 500, "internal server error", nil)
		return
	}

	// Map gRPC status codes to HTTP status codes.
	httpStatus := grpcCodeToHTTP(uint32(st.Code()))
	respond(c, httpStatus, int(st.Code()), st.Message(), nil)
}

func grpcCodeToHTTP(code uint32) int {
	switch code {
	case 0:
		return http.StatusOK
	case 3, 5, 9:
		return http.StatusBadRequest
	case 4:
		return http.StatusRequestTimeout
	case 6, 7:
		return http.StatusNotFound
	case 8, 10, 16:
		return http.StatusForbidden
	case 13:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
