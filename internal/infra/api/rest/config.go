package rest

import (
	"log"
	"time"
)

type ServerConfig struct {
	ErrorLog          *log.Logger
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}
