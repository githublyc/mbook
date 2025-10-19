package main

import (
	"mbook/webook/internal/events"
	"mbook/webook/pkg/grpcx"
)

type App struct {
	consumers []events.Consumer
	server    *grpcx.Server
}
