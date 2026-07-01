package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	pbOrder "github.com/storm/myidea/api/order/v1"
	pbPayment "github.com/storm/myidea/api/payment/v1"
	pbProduct "github.com/storm/myidea/api/product/v1"
	pbUser "github.com/storm/myidea/api/user/v1"
)

type AdminHandler struct {
	orderClient   pbOrder.OrderServiceClient
	productClient pbProduct.ProductServiceClient
	userClient    pbUser.UserServiceClient
	paymentClient pbPayment.PaymentServiceClient
}

func NewAdminHandler(orderClient pbOrder.OrderServiceClient, productClient pbProduct.ProductServiceClient, userClient pbUser.UserServiceClient, paymentClient pbPayment.PaymentServiceClient) *AdminHandler {
	return &AdminHandler{
		orderClient:   orderClient,
		productClient: productClient,
		userClient:    userClient,
		paymentClient: paymentClient,
	}
}

// Dashboard  GET /api/v1/admin/dashboard
func (h *AdminHandler) Dashboard(c *gin.Context) {
	ctx := c.Request.Context()

	type Stats struct {
		TotalOrders   int64   `json:"total_orders"`
		TotalRevenue  float64 `json:"total_revenue"`
		TotalUsers    int64   `json:"total_users"`
		TotalProducts int64   `json:"total_products"`
		RecentOrders  any     `json:"recent_orders"`
	}

	stats := Stats{}

	ordersResp, err := h.orderClient.AdminListOrders(ctx, &pbOrder.AdminListOrdersRequest{Page: 1, PageSize: 1})
	if err == nil {
		stats.TotalOrders = int64(ordersResp.Total)
	}

	allOrdersResp, err := h.orderClient.AdminListOrders(ctx, &pbOrder.AdminListOrdersRequest{Page: 1, PageSize: 1000})
	if err == nil {
		for _, o := range allOrdersResp.Orders {
			if o.Status == 1 || o.Status == 2 || o.Status == 3 {
				stats.TotalRevenue += o.TotalAmount
			}
		}
	}

	usersResp, err := h.userClient.AdminListUsers(ctx, &pbUser.AdminListUsersRequest{Page: 1, PageSize: 1})
	if err == nil {
		stats.TotalUsers = int64(usersResp.Total)
	}

	spuResp, err := h.productClient.ListSPUs(ctx, &pbProduct.ListSPUsRequest{Page: 1, PageSize: 1})
	if err == nil {
		stats.TotalProducts = int64(spuResp.Total)
	}

	recentResp, _ := h.orderClient.AdminListOrders(ctx, &pbOrder.AdminListOrdersRequest{Page: 1, PageSize: 10})
	if recentResp != nil {
		stats.RecentOrders = recentResp.Orders
	}

	respond(c, http.StatusOK, 0, "ok", stats)
}

// ListOrders  GET /api/v1/admin/orders
func (h *AdminHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	ctx := c.Request.Context()
	resp, err := h.orderClient.AdminListOrders(ctx, &pbOrder.AdminListOrdersRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Status:   int32(status),
		Keyword:  keyword,
		DateFrom: dateFrom,
		DateTo:   dateTo,
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

// GetOrder  GET /api/v1/admin/orders/:id
func (h *AdminHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid order id", nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.orderClient.GetOrder(ctx, &pbOrder.GetOrderRequest{Id: id})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{"order": resp.Order})
}

// ShipOrder  POST /api/v1/admin/orders/:id/ship
func (h *AdminHandler) ShipOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid order id", nil)
		return
	}

	ctx := c.Request.Context()
	_, err = h.orderClient.UpdateOrderStatus(ctx, &pbOrder.UpdateOrderStatusRequest{
		Id:     id,
		Status: 2,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "order shipped", nil)
}

// RefundOrder  POST /api/v1/admin/orders/:id/refund
func (h *AdminHandler) RefundOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid order id", nil)
		return
	}

	ctx := c.Request.Context()
	_, err = h.orderClient.UpdateOrderStatus(ctx, &pbOrder.UpdateOrderStatusRequest{
		Id:     id,
		Status: 5,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "order refunding", nil)
}

// CreateSPU  POST /api/v1/admin/products/spus
func (h *AdminHandler) CreateSPU(c *gin.Context) {
	var req struct {
		CategoryID         int64    `json:"category_id" binding:"required"`
		BrandID            int64    `json:"brand_id" binding:"required"`
		Title              string   `json:"title" binding:"required"`
		Subtitle           string   `json:"subtitle"`
		SaleableAttrNames  []string `json:"saleable_attr_names"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.productClient.CreateSPU(ctx, &pbProduct.CreateSPURequest{
		CategoryId:        req.CategoryID,
		BrandId:           req.BrandID,
		Title:             req.Title,
		Subtitle:          req.Subtitle,
		SaleableAttrNames: req.SaleableAttrNames,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Spu)
}

// UpdateSPU  PUT /api/v1/admin/products/spus/:id
func (h *AdminHandler) UpdateSPU(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid spu id", nil)
		return
	}

	var req struct {
		CategoryID        int64    `json:"category_id"`
		BrandID           int64    `json:"brand_id"`
		Title             string   `json:"title"`
		Subtitle          string   `json:"subtitle"`
		Status            int32    `json:"status"`
		SaleableAttrNames []string `json:"saleable_attr_names"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.productClient.UpdateSPU(ctx, &pbProduct.UpdateSPURequest{
		Id:                id,
		CategoryId:        req.CategoryID,
		BrandId:           req.BrandID,
		Title:             req.Title,
		Subtitle:          req.Subtitle,
		Status:            req.Status,
		SaleableAttrNames: req.SaleableAttrNames,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Spu)
}

// ListSPUs  GET /api/v1/admin/products/spus
func (h *AdminHandler) ListSPUs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	categoryID, _ := strconv.ParseInt(c.DefaultQuery("category_id", "0"), 10, 64)
	brandID, _ := strconv.ParseInt(c.DefaultQuery("brand_id", "0"), 10, 64)
	keyword := c.Query("keyword")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))

	ctx := c.Request.Context()
	resp, err := h.productClient.ListSPUs(ctx, &pbProduct.ListSPUsRequest{
		CategoryId: categoryID,
		BrandId:    brandID,
		Keyword:    keyword,
		Status:     int32(status),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{
		"spus":  resp.Spus,
		"total": resp.Total,
	})
}

// ListSKUs  GET /api/v1/admin/products/spus/:id/skus
func (h *AdminHandler) ListSKUs(c *gin.Context) {
	spuID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid spu id", nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.productClient.ListSKUs(ctx, &pbProduct.ListSKUsRequest{SpuId: spuID})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{"skus": resp.Skus})
}

// BatchCreateSKU  POST /api/v1/admin/products/skus
func (h *AdminHandler) BatchCreateSKU(c *gin.Context) {
	var req struct {
		SpuID int64 `json:"spu_id" binding:"required"`
		Skus  []struct {
			Attrs       map[string]string `json:"attrs"`
			Price       float64           `json:"price"`
			OriginPrice float64           `json:"origin_price"`
			Stock       int32             `json:"stock"`
			Code        string            `json:"code"`
			Image       string            `json:"image"`
		} `json:"skus" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	skuItems := make([]*pbProduct.CreateSKUItem, len(req.Skus))
	for i, s := range req.Skus {
		skuItems[i] = &pbProduct.CreateSKUItem{
			Attrs:       s.Attrs,
			Price:       s.Price,
			OriginPrice: s.OriginPrice,
			Stock:       s.Stock,
			Code:        s.Code,
			Image:       s.Image,
		}
	}

	ctx := c.Request.Context()
	resp, err := h.productClient.BatchCreateSKU(ctx, &pbProduct.BatchCreateSKURequest{
		SpuId: req.SpuID,
		Skus:  skuItems,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{"skus": resp.Skus})
}

// UpdateSKU  PUT /api/v1/admin/products/skus/:id
func (h *AdminHandler) UpdateSKU(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid sku id", nil)
		return
	}

	var req struct {
		Price       float64 `json:"price"`
		OriginPrice float64 `json:"origin_price"`
		Code        string  `json:"code"`
		Image       string  `json:"image"`
		Status      int32   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.productClient.UpdateSKU(ctx, &pbProduct.UpdateSKURequest{
		Id:          id,
		Price:       req.Price,
		OriginPrice: req.OriginPrice,
		Code:        req.Code,
		Image:       req.Image,
		Status:      req.Status,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Sku)
}

// CreateCategory  POST /api/v1/admin/products/categories
func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var req struct {
		ParentID  int64  `json:"parent_id"`
		Name      string `json:"name" binding:"required"`
		Icon      string `json:"icon"`
		SortOrder int32  `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.productClient.CreateCategory(ctx, &pbProduct.CreateCategoryRequest{
		ParentId:  req.ParentID,
		Name:      req.Name,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Category)
}

// UpdateCategory  PUT /api/v1/admin/products/categories/:id
func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid category id", nil)
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		Icon      string `json:"icon"`
		SortOrder int32  `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.productClient.UpdateCategory(ctx, &pbProduct.UpdateCategoryRequest{
		Id:        id,
		Name:      req.Name,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Category)
}

// DeleteCategory  DELETE /api/v1/admin/products/categories/:id
func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid category id", nil)
		return
	}

	ctx := c.Request.Context()
	_, err = h.productClient.DeleteCategory(ctx, &pbProduct.DeleteCategoryRequest{Id: id})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", nil)
}

// CreateBrand  POST /api/v1/admin/products/brands
func (h *AdminHandler) CreateBrand(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Logo      string `json:"logo"`
		SortOrder int32  `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.productClient.CreateBrand(ctx, &pbProduct.CreateBrandRequest{
		Name:      req.Name,
		Logo:      req.Logo,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp.Brand)
}

// ListUsers  GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	ctx := c.Request.Context()
	resp, err := h.userClient.AdminListUsers(ctx, &pbUser.AdminListUsersRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Keyword:  keyword,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{
		"users": resp.Users,
		"total": resp.Total,
	})
}

// ListPayments  GET /api/v1/admin/payments
func (h *AdminHandler) ListPayments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	orderNo := c.Query("order_no")

	ctx := c.Request.Context()
	resp, err := h.paymentClient.AdminListPayments(ctx, &pbPayment.AdminListPaymentsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Status:   int32(status),
		OrderNo:  orderNo,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{
		"payments": resp.Payments,
		"total":    resp.Total,
	})
}

// GetCategoryTree  GET /api/v1/admin/categories
func (h *AdminHandler) GetCategoryTree(c *gin.Context) {
	ctx := c.Request.Context()
	resp, err := h.productClient.GetCategoryTree(ctx, &pbProduct.GetCategoryTreeRequest{})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{"categories": resp.Categories})
}

// ListBrands  GET /api/v1/admin/brands
func (h *AdminHandler) ListBrands(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	ctx := c.Request.Context()
	resp, err := h.productClient.ListBrands(ctx, &pbProduct.ListBrandsRequest{
		Keyword:  keyword,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{
		"brands": resp.Brands,
		"total":  resp.Total,
	})
}

// createTime creates a time string for logging/audit purposes.
func createTime() string {
	return time.Now().Format(time.RFC3339)
}
