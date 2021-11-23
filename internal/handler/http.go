package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/docs"
	"github.com/iivkis/pos-ninja-backend/internal/myservice"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	docs.SwaggerInfo.BasePath = "/api/v1"

	root := h.engine.Group("/")

	root.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	root.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "POS-Ninja-Backend (version: 0.1-alpha)")
	})

	api := root.Group("/api")
	h.connectApiV1(api.Group("/v1"))

	return h.engine
}
