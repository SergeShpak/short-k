package rest

import (
	"log"
	"log/slog"
	"time"
)

//nolint:govet // fieldalignement check is irrelevant heree
type ServerConfig struct {
	ErrorLog *log.Logger
	ServerConfigParams
	Handler HandlerConfig
}

type ServerConfigParams struct {
	Host              string        `yaml:"host" validate:"required"`
	ReadTimeout       time.Duration `yaml:"readTimeout"`
	ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout"`
	WriteTimeout      time.Duration `yaml:"writeTimeout"`
	IdleTimeout       time.Duration `yaml:"idleTimeout"`
}

func GetDefaultServerConfigParams() ServerConfigParams {
	return ServerConfigParams{
		Host:              "",
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 1,
		WriteTimeout:      time.Second * 30,
		IdleTimeout:       time.Second * 120,
	}
}

type HandlerConfig struct {
	App    App
	Logger *slog.Logger
	HandlerConfigParams
}

type HandlerConfigParams struct {
	BaseAddr           string `yaml:"baseAddr" validate:"required,http_url"`
	MaxRequestBodySize int64  `yaml:"maxRequestBodySize" validate:"required,gt=0"`
}

func GetDefaultHandlerConfigParams() HandlerConfigParams {
	return HandlerConfigParams{
		BaseAddr:           "",
		MaxRequestBodySize: 8000,
	}
}
