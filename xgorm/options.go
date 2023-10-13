package xgorm

import (
	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

type Options struct {
	Logger                   gormlogger.Interface
	Models                   []interface{}
	AutoMigrate              bool
	Tracing                  bool
	Recreate                 bool
	SkipDefaultTransaction   bool
	DisableNestedTransaction bool
}

type Option func(*Options)

func WithModels(models ...interface{}) Option {
	return func(o *Options) {
		o.Models = append(o.Models, models...)
	}
}

func WithAutoMigrate() Option {
	return func(o *Options) {
		o.AutoMigrate = true
	}
}

func WithTracing() Option {
	return func(o *Options) {
		o.Tracing = true
	}
}

func WithLogger(z *zap.Logger, level gormlogger.LogLevel) Option {
	l := zapgorm2.
		New(z).
		LogMode(level)

	return func(o *Options) {
		o.Logger = l
	}
}

func WithRecreate() Option {
	return func(o *Options) {
		o.Recreate = true
	}
}

func WithSkipDefaultTransaction() Option {
	return func(o *Options) {
		o.SkipDefaultTransaction = true
	}
}

func WithDisableNestedTransaction() Option {
	return func(o *Options) {
		o.DisableNestedTransaction = true
	}
}

func NewOptions() *Options {
	return &Options{
		Models:      nil,
		AutoMigrate: false,
	}
}
