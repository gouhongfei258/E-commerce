package service

import (
	"context"

	pb "github.com/storm/myidea/api/cart/v1"
	"github.com/storm/myidea/service/cart/internal/biz"
)

type CartService struct {
	pb.UnimplementedCartServiceServer
	biz *biz.CartBiz
}

func NewCartService(biz *biz.CartBiz) *CartService {
	return &CartService{biz: biz}
}

func (s *CartService) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.CartItemProto, error) {
	item := &biz.CartItem{
		ProductID: req.ProductId,
		SKUID:     req.SkuId,
		SPUID:     req.SpuId,
		ShopID:    req.ShopId,
		Title:     req.Title,
		Image:     req.Image,
		Attrs:     req.Attrs,
		Price:     req.Price,
		Quantity:  req.Quantity,
	}

	if err := s.biz.AddItem(ctx, req.UserId, item); err != nil {
		return nil, err
	}
	return itemToProto(item), nil
}

func (s *CartService) UpdateQuantity(ctx context.Context, req *pb.UpdateQuantityRequest) (*pb.CartItemProto, error) {
	if err := s.biz.UpdateQuantity(ctx, req.UserId, req.ItemId, req.Quantity); err != nil {
		return nil, err
	}
	return &pb.CartItemProto{
		Id:       req.ItemId,
		UserId:   req.UserId,
		Quantity: req.Quantity,
	}, nil
}

func (s *CartService) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.Empty, error) {
	if err := s.biz.RemoveItem(ctx, req.UserId, req.ItemId); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *CartService) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	items, err := s.biz.ListItems(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	protos := make([]*pb.CartItemProto, len(items))
	for i, item := range items {
		protos[i] = itemToProto(item)
	}
	return &pb.ListItemsResponse{Items: protos}, nil
}

func (s *CartService) ClearItems(ctx context.Context, req *pb.ClearItemsRequest) (*pb.Empty, error) {
	if err := s.biz.ClearItems(ctx, req.UserId, req.ItemIds); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func itemToProto(item *biz.CartItem) *pb.CartItemProto {
	var createdAt, updatedAt string
	if !item.CreatedAt.IsZero() {
		createdAt = item.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !item.UpdatedAt.IsZero() {
		updatedAt = item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return &pb.CartItemProto{
		Id:        item.ID,
		UserId:    item.UserID,
		ProductId: item.ProductID,
		SkuId:     item.SKUID,
		SpuId:     item.SPUID,
		ShopId:    item.ShopID,
		Title:     item.Title,
		Image:     item.Image,
		Attrs:     item.Attrs,
		Price:     item.Price,
		Quantity:  item.Quantity,
		Selected:  item.Selected,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
