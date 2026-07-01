package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kratos/kratos/v2"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pbOrder "github.com/storm/myidea/api/order/v1"
	"github.com/storm/myidea/service/order/internal/adapter"
	"github.com/storm/myidea/service/order/internal/biz"
	"github.com/storm/myidea/service/order/internal/conf"
	"github.com/storm/myidea/service/order/internal/data"
	"github.com/storm/myidea/service/order/internal/service"
)

func main() {
	cfg := conf.LoadConfig()
	logger := klog.NewStdLogger(os.Stdout)
	klog.SetLogger(logger)

	dataObj, cleanup, err := data.NewData(cfg, logger)
	if err != nil {
		log.Fatalf("failed to init data: %v", err)
	}
	defer cleanup()

	productStockClient, err := data.NewProductStockClient(cfg.Server.GRPC.ProductAddr)
	if err != nil {
		log.Fatalf("failed to create product stock client: %v", err)
	}
	defer func() {
		if cl, ok := productStockClient.(interface{ Close() error }); ok {
			cl.Close()
		}
	}()

	orderRepo := data.NewOrderRepo(dataObj)
	orderBiz := biz.NewOrderBiz(orderRepo, productStockClient)
	orderSvc := service.NewOrderService(orderBiz)

	srv := grpc.NewServer()
	pbOrder.RegisterOrderServiceServer(srv, orderSvc)
	reflection.Register(srv)

	app := kratos.New(
		kratos.Name(cfg.Server.Name),
		kratos.Version(cfg.Server.Version),
		kratos.Server(adapter.GRPC{Srv: srv, Addr: cfg.Server.GRPC.Addr}),
	)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		srv.GracefulStop()
		app.Stop()
	}()

	if err := app.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
