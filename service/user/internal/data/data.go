package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/storm/myidea/service/user/internal/conf"
)

type Data struct {
	db *gorm.DB
}

type kratosWriter struct {
	logger *log.Helper
}

func (w *kratosWriter) Printf(format string, args ...any) {
	w.logger.Infof(format, args...)
}

func NewData(c *conf.Config, l log.Logger) (*Data, func(), error) {
	gormLogger := logger.New(
		&kratosWriter{logger: log.NewHelper(l)},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: c.Database.DSN,
	}), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open gorm db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(c.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.Database.ConnMaxLifetime)

	if err := db.AutoMigrate(
		&GORMUser{}, &GORMAddress{},
	); err != nil {
		return nil, nil, fmt.Errorf("auto migrate: %w", err)
	}

	d := &Data{db: db}
	cleanup := func() {
		sqlDB.Close()
	}
	return d, cleanup, nil
}

func (d *Data) DB(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}
