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

	pbUser "github.com/storm/myidea/api/user/v1"
	"github.com/storm/myidea/service/user/internal/adapter"
	"github.com/storm/myidea/service/user/internal/biz"
	"github.com/storm/myidea/service/user/internal/conf"
	"github.com/storm/myidea/service/user/internal/data"
	"github.com/storm/myidea/service/user/internal/service"
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

	userRepo := data.NewUserRepo(dataObj)
	userBiz := biz.NewUserBiz(userRepo)
	addressRepo := data.NewAddressRepo(dataObj)
	addressBiz := biz.NewAddressBiz(addressRepo)
	userSvc := service.NewUserService(userBiz, addressBiz)

	srv := grpc.NewServer()
	pbUser.RegisterUserServiceServer(srv, userSvc)
	reflection.Register(srv)

	app := kratos.New(
		kratos.Name("user-service"),
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
