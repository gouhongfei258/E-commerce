package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbCart "github.com/storm/myidea/api/cart/v1"
	pbOrder "github.com/storm/myidea/api/order/v1"
	pbPayment "github.com/storm/myidea/api/payment/v1"
)

// CartHandler handles HTTP requests for the cart domain.
type CartHandler struct {
	cartClient    pbCart.CartServiceClient
	orderClient   pbOrder.OrderServiceClient   // used by checkout
	paymentClient pbPayment.PaymentServiceClient // used by checkout
	conn          *grpc.ClientConn
}

// NewCartHandler creates a new handler and dials the cart gRPC endpoint.
// orderClient and paymentClient are used during checkout.
func NewCartHandler(grpcAddr string, dialTimeout time.Duration,
	orderClient pbOrder.OrderServiceClient,
	paymentClient pbPayment.PaymentServiceClient,
) (*CartHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryClientInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	return &CartHandler{
		cartClient:    pbCart.NewCartServiceClient(conn),
		orderClient:   orderClient,
		paymentClient: paymentClient,
		conn:          conn,
	}, nil
}

// Close shuts down the gRPC connection.
func (h *CartHandler) Close() error {
	return h.conn.Close()
}

// ListItems  GET /api/v1/cart
func (h *CartHandler) ListItems(c *gin.Context) {
	userID := c.GetInt64("user_id")
	ctx := injectUserID(c, userID)

	resp, err := h.cartClient.ListItems(ctx, &pbCart.ListItemsRequest{UserId: userID})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{"items": resp.Items})
}

// AddItem  POST /api/v1/cart/items
func (h *CartHandler) AddItem(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ProductID int64             `json:"product_id" binding:"required"`
		SKUID     int64             `json:"sku_id" binding:"required"`
		SPUID     int64             `json:"spu_id" binding:"required"`
		ShopID    int64             `json:"shop_id"`
		Title     string            `json:"title" binding:"required"`
		Image     string            `json:"image"`
		Attrs     map[string]string `json:"attrs"`
		Price     float64           `json:"price" binding:"required,gt=0"`
		Quantity  int32             `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := injectUserID(c, userID)
	resp, err := h.cartClient.AddItem(ctx, &pbCart.AddItemRequest{
		UserId:    userID,
		ProductId: req.ProductID,
		SkuId:     req.SKUID,
		SpuId:     req.SPUID,
		ShopId:    req.ShopID,
		Title:     req.Title,
		Image:     req.Image,
		Attrs:     req.Attrs,
		Price:     req.Price,
		Quantity:  req.Quantity,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp)
}

// UpdateQuantity  PUT /api/v1/cart/items/:id
func (h *CartHandler) UpdateQuantity(c *gin.Context) {
	userID := c.GetInt64("user_id")
	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid item id", nil)
		return
	}

	var req struct {
		Quantity int32 `json:"quantity" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := injectUserID(c, userID)
	resp, err := h.cartClient.UpdateQuantity(ctx, &pbCart.UpdateQuantityRequest{
		UserId:   userID,
		ItemId:   itemID,
		Quantity: req.Quantity,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp)
}

// RemoveItem  DELETE /api/v1/cart/items/:id
func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID := c.GetInt64("user_id")
	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid item id", nil)
		return
	}

	ctx := injectUserID(c, userID)
	_, err = h.cartClient.RemoveItem(ctx, &pbCart.RemoveItemRequest{
		UserId: userID,
		ItemId: itemID,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", nil)
}

// Checkout  POST /api/v1/cart/checkout
// Lists the current user's cart items, creates an order (with stock locking),
// initiates payment, and processes the result.
func (h *CartHandler) Checkout(c *gin.Context) {
	userID := c.GetInt64("user_id")

	if h.orderClient == nil {
		respond(c, http.StatusInternalServerError, 500, "order service not available", nil)
		return
	}
	if h.paymentClient == nil {
		respond(c, http.StatusInternalServerError, 500, "payment service not available", nil)
		return
	}

	ctx := injectUserID(c, userID)

	// 1. List cart items.
	cartResp, err := h.cartClient.ListItems(ctx, &pbCart.ListItemsRequest{UserId: userID})
	if err != nil {
		respondError(c, err)
		return
	}

	if len(cartResp.Items) == 0 {
		respond(c, http.StatusBadRequest, 400, "cart is empty", nil)
		return
	}

	// 2. Build order items from cart (include SKU ID for stock locking).
	var req struct {
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

	orderItems := make([]*pbOrder.OrderItemProto, 0, len(cartResp.Items))
	itemIDs := make([]int64, 0, len(cartResp.Items))
	for _, item := range cartResp.Items {
		orderItems = append(orderItems, &pbOrder.OrderItemProto{
			SkuId:       item.SkuId,
			ProductId:   item.ProductId,
			ProductName: item.Title,
			Image:       item.Image,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
		itemIDs = append(itemIDs, item.Id)
	}

	// 3. Create order (Order Service internally locks stock).
	orderResp, err := h.orderClient.CreateOrder(ctx, &pbOrder.CreateOrderRequest{
		UserId: userID,
		Items:  orderItems,
		Address: &pbOrder.ShippingAddressProto{
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

	// 4. Clear checked-out items.
	_, _ = h.cartClient.ClearItems(ctx, &pbCart.ClearItemsRequest{
		UserId:  userID,
		ItemIds: itemIDs,
	})

	// 5. Create payment for the order.
	paymentResp, err := h.paymentClient.CreatePayment(ctx, &pbPayment.CreatePaymentRequest{
		UserId:      userID,
		OrderNo:     orderResp.Order.OrderNo,
		TotalAmount: orderResp.Order.TotalAmount,
		Method:      pbPayment.PaymentMethod_MOCK,
	})
	if err != nil {
		// Payment creation failed — cancel order to unlock stock.
		_, _ = h.orderClient.CancelOrder(ctx, &pbOrder.CancelOrderRequest{
			Id:     orderResp.Order.Id,
			UserId: userID,
		})
		respondError(c, err)
		return
	}

	// 6. Process payment (mock: 80% success / 20% failure).
	processResp, err := h.paymentClient.ProcessPayment(ctx, &pbPayment.ProcessPaymentRequest{
		PaymentId: paymentResp.Payment.Id,
		Method:    pbPayment.PaymentMethod_MOCK,
	})
	if err != nil || processResp.Payment.Status != pbPayment.PaymentStatus_SUCCESS {
		// Payment failed — cancel order to unlock stock.
		_, _ = h.orderClient.CancelOrder(ctx, &pbOrder.CancelOrderRequest{
			Id:     orderResp.Order.Id,
			UserId: userID,
		})
		respond(c, http.StatusOK, 0, "payment failed, order cancelled", gin.H{
			"order":   orderResp.Order,
			"payment": processResp.GetPayment(),
		})
		return
	}

	// 7. Payment succeeded — confirm order (triggers ConfirmDeductStock).
	_, err = h.orderClient.UpdateOrderStatus(ctx, &pbOrder.UpdateOrderStatusRequest{
		Id:     orderResp.Order.Id,
		Status: 1, // OrderStatusPaid
	})
	if err != nil {
		respondError(c, err)
		return
	}

	orderResp.Order.Status = 1 // Paid

	respond(c, http.StatusOK, 0, "checkout completed", gin.H{
		"order":   orderResp.Order,
		"payment": processResp.Payment,
	})
}
