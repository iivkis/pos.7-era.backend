package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/myservice"
)

type HttpHandler struct {
	engine  *gin.Engine
	service myservice.MyService
}

func NewHttpHandler(service myservice.MyService) HttpHandler {
	gin.SetMode(gin.ReleaseMode)

	//create engine
	engine := gin.Default()
	engine.Use(gin.Recovery(), gin.Logger())

	return HttpHandler{
		engine:  engine,
		service: service,
	}
}

func (h *HttpHandler) Init() *gin.Engine {
	root := h.engine.Group("/")
	api := root.Group("/api")

	h.connectApiV1(api.Group("/v1"))

	return h.engine
}
