//пакет инициализирует swagger документацию в переданном в движке

package swagdocs

import (
	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(engine *gin.Engine) {
	docs.SwaggerInfo.BasePath = "/api/v1"
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
