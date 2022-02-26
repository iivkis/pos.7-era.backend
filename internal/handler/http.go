package handler

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/docs"
	"github.com/iivkis/pos-ninja-backend/internal/myservice"
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type HttpHandler struct {
	engine  *gin.Engine
	srv     myservice.MyService
	authjwt *authjwt.AuthJWT
}

func NewHttpHandler(service myservice.MyService, authjwt *authjwt.AuthJWT) HttpHandler {
	gin.SetMode(gin.ReleaseMode)

	//create engine
	engine := gin.Default()

	//set up recovery and logger
	engine.Use(gin.Recovery(), gin.Logger())

	//use CORS
	engine.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		MaxAge:           12 * time.Hour,
	}))

	return HttpHandler{
		engine:  engine,
		srv:     service,
		authjwt: authjwt,
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
