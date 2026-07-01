package service

import (
	"context"

	pb "github.com/storm/myidea/api/product/v1"
	"github.com/storm/myidea/service/product/internal/biz"
)

type ProductService struct {
	pb.UnimplementedProductServiceServer
	categoryBiz *biz.CategoryBiz
	brandBiz    *biz.BrandBiz
	spuBiz      *biz.SPUBiz
	skuBiz      *biz.SKUBiz
}

func NewProductService(categoryBiz *biz.CategoryBiz, brandBiz *biz.BrandBiz, spuBiz *biz.SPUBiz, skuBiz *biz.SKUBiz) *ProductService {
	return &ProductService{
		categoryBiz: categoryBiz,
		brandBiz:    brandBiz,
		spuBiz:      spuBiz,
		skuBiz:      skuBiz,
	}
}

func (s *ProductService) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CategoryResponse, error) {
	cat, err := s.categoryBiz.CreateCategory(ctx, req.Name, req.ParentId, req.Icon, req.SortOrder)
	if err != nil {
		return nil, err
	}
	return &pb.CategoryResponse{Category: categoryToProto(cat)}, nil
}

func (s *ProductService) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.CategoryResponse, error) {
	cat, err := s.categoryBiz.UpdateCategory(ctx, req.Id, req.Name, req.Icon, req.SortOrder)
	if err != nil {
		return nil, err
	}
	return &pb.CategoryResponse{Category: categoryToProto(cat)}, nil
}

func (s *ProductService) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.Empty, error) {
	if err := s.categoryBiz.DeleteCategory(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *ProductService) GetCategoryTree(ctx context.Context, req *pb.GetCategoryTreeRequest) (*pb.GetCategoryTreeResponse, error) {
	tree, err := s.categoryBiz.GetCategoryTree(ctx)
	if err != nil {
		return nil, err
	}
	protos := make([]*pb.CategoryProto, len(tree))
	for i, c := range tree {
		protos[i] = categoryToProto(c)
	}
	return &pb.GetCategoryTreeResponse{Categories: protos}, nil
}

func (s *ProductService) CreateBrand(ctx context.Context, req *pb.CreateBrandRequest) (*pb.BrandResponse, error) {
	brand, err := s.brandBiz.CreateBrand(ctx, req.Name, req.Logo, req.SortOrder)
	if err != nil {
		return nil, err
	}
	return &pb.BrandResponse{Brand: brandToProto(brand)}, nil
}

func (s *ProductService) ListBrands(ctx context.Context, req *pb.ListBrandsRequest) (*pb.ListBrandsResponse, error) {
	brands, total, err := s.brandBiz.ListBrands(ctx, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	protos := make([]*pb.BrandProto, len(brands))
	for i, b := range brands {
		protos[i] = brandToProto(b)
	}
	return &pb.ListBrandsResponse{Brands: protos, Total: total}, nil
}

func (s *ProductService) CreateSPU(ctx context.Context, req *pb.CreateSPURequest) (*pb.SPUResponse, error) {
	spu, err := s.spuBiz.CreateSPU(ctx, req.CategoryId, req.BrandId, req.Title, req.Subtitle, req.SaleableAttrNames)
	if err != nil {
		return nil, err
	}
	return &pb.SPUResponse{Spu: spuToProto(spu, nil)}, nil
}

func (s *ProductService) UpdateSPU(ctx context.Context, req *pb.UpdateSPURequest) (*pb.SPUResponse, error) {
	spu := &biz.SPU{
		ID:               req.Id,
		CategoryID:       req.CategoryId,
		BrandID:          req.BrandId,
		Title:            req.Title,
		Subtitle:         req.Subtitle,
		Status:           biz.SPUStatus(req.Status),
		SaleableAttrNames: req.SaleableAttrNames,
	}
	result, err := s.spuBiz.UpdateSPU(ctx, spu)
	if err != nil {
		return nil, err
	}
	return &pb.SPUResponse{Spu: spuToProto(result, nil)}, nil
}

func (s *ProductService) GetSPU(ctx context.Context, req *pb.GetSPURequest) (*pb.SPUResponse, error) {
	spu, err := s.spuBiz.GetSPU(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.SPUResponse{Spu: spuToProto(spu, nil)}, nil
}

func (s *ProductService) ListSPUs(ctx context.Context, req *pb.ListSPUsRequest) (*pb.ListSPUsResponse, error) {
	filter := &biz.SPUFilter{
		CategoryID: req.CategoryId,
		BrandID:    req.BrandId,
		Keyword:    req.Keyword,
		Status:     biz.SPUStatus(req.Status),
	}
	spus, total, err := s.spuBiz.ListSPUs(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	protos := make([]*pb.SPUProto, len(spus))
	for i, sp := range spus {
		protos[i] = spuToProto(sp, nil)
	}
	return &pb.ListSPUsResponse{Spus: protos, Total: total}, nil
}

func (s *ProductService) BatchCreateSKU(ctx context.Context, req *pb.BatchCreateSKURequest) (*pb.BatchCreateSKUResponse, error) {
	skus := make([]*biz.SKU, len(req.Skus))
	for i, item := range req.Skus {
		skus[i] = &biz.SKU{
			Attrs:       item.Attrs,
			Price:       item.Price,
			OriginPrice: item.OriginPrice,
			Stock:       item.Stock,
			Code:        item.Code,
			Image:       item.Image,
		}
	}
	result, err := s.skuBiz.BatchCreateSKU(ctx, req.SpuId, skus)
	if err != nil {
		return nil, err
	}
	protos := make([]*pb.SKUProto, len(result))
	for i, sk := range result {
		protos[i] = skuToProto(sk)
	}
	return &pb.BatchCreateSKUResponse{Skus: protos}, nil
}

func (s *ProductService) UpdateSKU(ctx context.Context, req *pb.UpdateSKURequest) (*pb.SKUResponse, error) {
	sku := &biz.SKU{
		ID:          req.Id,
		Price:       req.Price,
		OriginPrice: req.OriginPrice,
		Code:        req.Code,
		Image:       req.Image,
		Status:      req.Status,
	}
	if err := s.skuBiz.UpdateSKU(ctx, sku); err != nil {
		return nil, err
	}
	return &pb.SKUResponse{Sku: skuToProto(sku)}, nil
}

func (s *ProductService) ListSKUs(ctx context.Context, req *pb.ListSKUsRequest) (*pb.ListSKUsResponse, error) {
	skus, err := s.skuBiz.ListSKUs(ctx, req.SpuId)
	if err != nil {
		return nil, err
	}
	protos := make([]*pb.SKUProto, len(skus))
	for i, sk := range skus {
		protos[i] = skuToProto(sk)
	}
	return &pb.ListSKUsResponse{Skus: protos}, nil
}

func (s *ProductService) LockStock(ctx context.Context, req *pb.LockStockRequest) (*pb.LockStockResponse, error) {
	for _, item := range req.Items {
		if err := s.skuBiz.LockStock(ctx, item.SkuId, item.Quantity, req.OrderNo); err != nil {
			return &pb.LockStockResponse{Success: false}, err
		}
	}
	return &pb.LockStockResponse{Success: true}, nil
}

func (s *ProductService) ConfirmDeductStock(ctx context.Context, req *pb.ConfirmDeductRequest) (*pb.Empty, error) {
	for _, item := range req.Items {
		if err := s.skuBiz.ConfirmDeductStock(ctx, item.SkuId, item.Quantity, req.OrderNo); err != nil {
			return nil, err
		}
	}
	return &pb.Empty{}, nil
}

func (s *ProductService) UnlockStock(ctx context.Context, req *pb.UnlockStockRequest) (*pb.Empty, error) {
	for _, item := range req.Items {
		if err := s.skuBiz.UnlockStock(ctx, item.SkuId, item.Quantity, req.OrderNo); err != nil {
			return nil, err
		}
	}
	return &pb.Empty{}, nil
}

func categoryToProto(c *biz.Category) *pb.CategoryProto {
	p := &pb.CategoryProto{
		Id:        c.ID,
		ParentId:  c.ParentID,
		Name:      c.Name,
		Icon:      c.Icon,
		SortOrder: c.SortOrder,
		Level:     c.Level,
	}
	if !c.CreatedAt.IsZero() {
		p.CreatedAt = c.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	for _, child := range c.Children {
		p.Children = append(p.Children, categoryToProto(child))
	}
	return p
}

func brandToProto(b *biz.Brand) *pb.BrandProto {
	p := &pb.BrandProto{
		Id:        b.ID,
		Name:      b.Name,
		Logo:      b.Logo,
		SortOrder: b.SortOrder,
	}
	if !b.CreatedAt.IsZero() {
		p.CreatedAt = b.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return p
}

func spuToProto(sp *biz.SPU, skus []*biz.SKU) *pb.SPUProto {
	p := &pb.SPUProto{
		Id:               sp.ID,
		CategoryId:       sp.CategoryID,
		BrandId:          sp.BrandID,
		Title:            sp.Title,
		Subtitle:         sp.Subtitle,
		Status:           int32(sp.Status),
		SaleableAttrNames: sp.SaleableAttrNames,
		SaleCount:        sp.SaleCount,
	}
	if !sp.CreatedAt.IsZero() {
		p.CreatedAt = sp.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !sp.UpdatedAt.IsZero() {
		p.UpdatedAt = sp.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	for _, sk := range skus {
		p.Skus = append(p.Skus, skuToProto(sk))
	}
	return p
}

func skuToProto(sk *biz.SKU) *pb.SKUProto {
	p := &pb.SKUProto{
		Id:          sk.ID,
		SpuId:       sk.SPUID,
		Attrs:       sk.Attrs,
		Price:       sk.Price,
		OriginPrice: sk.OriginPrice,
		Stock:       sk.Stock,
		LockedStock: sk.LockedStock,
		Code:        sk.Code,
		Image:       sk.Image,
		Status:      sk.Status,
		SaleCount:   sk.SaleCount,
		Version:     sk.Version,
	}
	if !sk.CreatedAt.IsZero() {
		p.CreatedAt = sk.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !sk.UpdatedAt.IsZero() {
		p.UpdatedAt = sk.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return p
}

var _ pb.ProductServiceServer = (*ProductService)(nil)
