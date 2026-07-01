package data

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbProduct "github.com/storm/myidea/api/product/v1"
	"github.com/storm/myidea/service/order/internal/biz"
)

type productStockClient struct {
	client pbProduct.ProductServiceClient
	conn   *grpc.ClientConn
}

func NewProductStockClient(grpcAddr string) (biz.ProductStockClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial product service: %w", err)
	}

	return &productStockClient{
		client: pbProduct.NewProductServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *productStockClient) Close() error {
	return c.conn.Close()
}

func (c *productStockClient) LockStock(ctx context.Context, skuID int64, quantity int32, orderNo string) error {
	resp, err := c.client.LockStock(ctx, &pbProduct.LockStockRequest{
		OrderNo: orderNo,
		Items: []*pbProduct.LockStockItem{
			{SkuId: skuID, Quantity: quantity},
		},
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return biz.ErrStockInsufficient
	}
	return nil
}

func (c *productStockClient) ConfirmDeductStock(ctx context.Context, skuID int64, quantity int32, orderNo string) error {
	_, err := c.client.ConfirmDeductStock(ctx, &pbProduct.ConfirmDeductRequest{
		OrderNo: orderNo,
		Items: []*pbProduct.ConfirmDeductItem{
			{SkuId: skuID, Quantity: quantity},
		},
	})
	return err
}

func (c *productStockClient) UnlockStock(ctx context.Context, skuID int64, quantity int32, orderNo string) error {
	_, err := c.client.UnlockStock(ctx, &pbProduct.UnlockStockRequest{
		OrderNo: orderNo,
		Items: []*pbProduct.UnlockStockItem{
			{SkuId: skuID, Quantity: quantity},
		},
	})
	return err
}
