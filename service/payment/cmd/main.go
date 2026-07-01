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

	pbPayment "github.com/storm/myidea/api/payment/v1"
	"github.com/storm/myidea/service/payment/internal/adapter"
	"github.com/storm/myidea/service/payment/internal/biz"
	"github.com/storm/myidea/service/payment/internal/conf"
	"github.com/storm/myidea/service/payment/internal/data"
	"github.com/storm/myidea/service/payment/internal/service"
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

	paymentRepo := data.NewPaymentRepo(dataObj)
	providerFactory := biz.NewDefaultPaymentProviderFactory()
	providerFactory.Register(biz.PaymentMethodMock, data.NewMockPaymentProvider())
	paymentBiz := biz.NewPaymentBiz(paymentRepo, providerFactory)
	paymentSvc := service.NewPaymentService(paymentBiz)

	srv := grpc.NewServer()
	pbPayment.RegisterPaymentServiceServer(srv, paymentSvc)
	reflection.Register(srv)

	app := kratos.New(
		kratos.Name("payment-service"),
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
