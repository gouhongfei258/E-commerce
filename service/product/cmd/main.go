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

	pbProduct "github.com/storm/myidea/api/product/v1"
	"github.com/storm/myidea/service/product/internal/adapter"
	"github.com/storm/myidea/service/product/internal/biz"
	"github.com/storm/myidea/service/product/internal/conf"
	"github.com/storm/myidea/service/product/internal/data"
	"github.com/storm/myidea/service/product/internal/service"
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

	categoryRepo := data.NewCategoryRepo(dataObj)
	categoryBiz := biz.NewCategoryBiz(categoryRepo)
	brandRepo := data.NewBrandRepo(dataObj)
	brandBiz := biz.NewBrandBiz(brandRepo)
	spuRepo := data.NewSPURepo(dataObj)
	spuBiz := biz.NewSPUBiz(spuRepo)
	skuRepo := data.NewSKURepo(dataObj)
	skuBiz := biz.NewSKUBiz(skuRepo)
	productSvc := service.NewProductService(categoryBiz, brandBiz, spuBiz, skuBiz)

	srv := grpc.NewServer()
	pbProduct.RegisterProductServiceServer(srv, productSvc)
	reflection.Register(srv)

	app := kratos.New(
		kratos.Name("product-service"),
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
