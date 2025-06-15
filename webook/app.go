package main

import (
	"github.com/gin-gonic/gin"
	"mbook/webook/events"
)

type App struct {
	server    *gin.Engine
	consumers []events.Consumer
}
