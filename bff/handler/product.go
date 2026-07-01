package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/storm/myidea/api/product/v1"
)

// ProductHandler handles HTTP requests for the product domain.
type ProductHandler struct {
	client pb.ProductServiceClient
	conn   *grpc.ClientConn
}

// NewProductHandler creates a new handler and dials the gRPC connection.
func NewProductHandler(grpcAddr string, dialTimeout time.Duration) (*ProductHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryClientInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	return &ProductHandler{
		client: pb.NewProductServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close shuts down the gRPC connection.
func (h *ProductHandler) Close() error {
	return h.conn.Close()
}

func (h *ProductHandler) ProductServiceClient() pb.ProductServiceClient {
	return h.client
}

// ------------------ Category ------------------

// GetCategoryTree  GET /api/v1/categories
func (h *ProductHandler) GetCategoryTree(c *gin.Context) {
	ctx := c.Request.Context()
	resp, err := h.client.GetCategoryTree(ctx, &pb.GetCategoryTreeRequest{})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", resp.Categories)
}

// CreateCategory  POST /api/v1/categories
func (h *ProductHandler) CreateCategory(c *gin.Context) {
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
	resp, err := h.client.CreateCategory(ctx, &pb.CreateCategoryRequest{
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

// UpdateCategory  PUT /api/v1/categories/:id
func (h *ProductHandler) UpdateCategory(c *gin.Context) {
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
	resp, err := h.client.UpdateCategory(ctx, &pb.UpdateCategoryRequest{
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

// DeleteCategory  DELETE /api/v1/categories/:id
func (h *ProductHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid category id", nil)
		return
	}

	ctx := c.Request.Context()
	_, err = h.client.DeleteCategory(ctx, &pb.DeleteCategoryRequest{Id: id})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", nil)
}

// ------------------ Brand ------------------

// ListBrands  GET /api/v1/brands
func (h *ProductHandler) ListBrands(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	ctx := c.Request.Context()
	resp, err := h.client.ListBrands(ctx, &pb.ListBrandsRequest{
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

// CreateBrand  POST /api/v1/brands
func (h *ProductHandler) CreateBrand(c *gin.Context) {
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
	resp, err := h.client.CreateBrand(ctx, &pb.CreateBrandRequest{
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

// ------------------ SPU ------------------

// ListSPUs  GET /api/v1/spus
func (h *ProductHandler) ListSPUs(c *gin.Context) {
	categoryID, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	brandID, _ := strconv.ParseInt(c.Query("brand_id"), 10, 64)
	keyword := c.Query("keyword")
	status, _ := strconv.Atoi(c.Query("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	ctx := c.Request.Context()
	resp, err := h.client.ListSPUs(ctx, &pb.ListSPUsRequest{
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

// GetSPU  GET /api/v1/spus/:id
func (h *ProductHandler) GetSPU(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid spu id", nil)
		return
	}

	ctx := c.Request.Context()
	spuResp, err := h.client.GetSPU(ctx, &pb.GetSPURequest{Id: id})
	if err != nil {
		respondError(c, err)
		return
	}

	// Also fetch SKUs for this SPU.
	skuResp, err := h.client.ListSKUs(ctx, &pb.ListSKUsRequest{SpuId: id})
	if err == nil {
		spuResp.Spu.Skus = skuResp.Skus
	}

	respond(c, http.StatusOK, 0, "ok", spuResp.Spu)
}

// CreateSPU  POST /api/v1/spus
func (h *ProductHandler) CreateSPU(c *gin.Context) {
	var req struct {
		CategoryID       int64    `json:"category_id" binding:"required"`
		BrandID          int64    `json:"brand_id" binding:"required"`
		Title            string   `json:"title" binding:"required"`
		Subtitle         string   `json:"subtitle"`
		SaleableAttrNames []string `json:"saleable_attr_names"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.client.CreateSPU(ctx, &pb.CreateSPURequest{
		CategoryId:       req.CategoryID,
		BrandId:          req.BrandID,
		Title:            req.Title,
		Subtitle:         req.Subtitle,
		SaleableAttrNames: req.SaleableAttrNames,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", resp.Spu)
}

// UpdateSPU  PUT /api/v1/spus/:id
func (h *ProductHandler) UpdateSPU(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid spu id", nil)
		return
	}

	var req struct {
		CategoryID       int64    `json:"category_id"`
		BrandID          int64    `json:"brand_id"`
		Title            string   `json:"title"`
		Subtitle         string   `json:"subtitle"`
		Status           int32    `json:"status"`
		SaleableAttrNames []string `json:"saleable_attr_names"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.client.UpdateSPU(ctx, &pb.UpdateSPURequest{
		Id:               id,
		CategoryId:       req.CategoryID,
		BrandId:          req.BrandID,
		Title:            req.Title,
		Subtitle:         req.Subtitle,
		Status:           req.Status,
		SaleableAttrNames: req.SaleableAttrNames,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", resp.Spu)
}

// UpdateSPUStatus  PUT /api/v1/spus/:id/status
func (h *ProductHandler) UpdateSPUStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid spu id", nil)
		return
	}

	var req struct {
		Status int32 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	spuResp, err := h.client.GetSPU(ctx, &pb.GetSPURequest{Id: id})
	if err != nil {
		respondError(c, err)
		return
	}

	resp, err := h.client.UpdateSPU(ctx, &pb.UpdateSPURequest{
		Id:               id,
		CategoryId:       spuResp.Spu.CategoryId,
		BrandId:          spuResp.Spu.BrandId,
		Title:            spuResp.Spu.Title,
		Subtitle:         spuResp.Spu.Subtitle,
		Status:           req.Status,
		SaleableAttrNames: spuResp.Spu.SaleableAttrNames,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", resp.Spu)
}

// ------------------ SKU ------------------

// ListSKUs  GET /api/v1/skus?spu_id=
func (h *ProductHandler) ListSKUs(c *gin.Context) {
	spuID, _ := strconv.ParseInt(c.Query("spu_id"), 10, 64)

	ctx := c.Request.Context()
	resp, err := h.client.ListSKUs(ctx, &pb.ListSKUsRequest{SpuId: spuID})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", resp.Skus)
}

// BatchCreateSKU  POST /api/v1/skus/batch
func (h *ProductHandler) BatchCreateSKU(c *gin.Context) {
	var req struct {
		SPUID int64 `json:"spu_id" binding:"required"`
		SKUs  []struct {
			Attrs       map[string]string `json:"attrs"`
			Price       float64           `json:"price" binding:"required,gt=0"`
			OriginPrice float64           `json:"origin_price"`
			Stock       int32             `json:"stock"`
			Code        string            `json:"code"`
			Image       string            `json:"image"`
		} `json:"skus" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	items := make([]*pb.CreateSKUItem, len(req.SKUs))
	for i, sk := range req.SKUs {
		items[i] = &pb.CreateSKUItem{
			Attrs:       sk.Attrs,
			Price:       sk.Price,
			OriginPrice: sk.OriginPrice,
			Stock:       sk.Stock,
			Code:        sk.Code,
			Image:       sk.Image,
		}
	}

	ctx := c.Request.Context()
	resp, err := h.client.BatchCreateSKU(ctx, &pb.BatchCreateSKURequest{
		SpuId: req.SPUID,
		Skus:  items,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, http.StatusOK, 0, "ok", resp.Skus)
}

// UpdateSKU  PUT /api/v1/skus/:id
func (h *ProductHandler) UpdateSKU(c *gin.Context) {
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
	resp, err := h.client.UpdateSKU(ctx, &pb.UpdateSKURequest{
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
