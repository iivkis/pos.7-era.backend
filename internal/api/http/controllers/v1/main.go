package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"github.com/iivkis/pos.7-era.backend/pkg/mailagent"
	"github.com/iivkis/strcode"
)

type combine struct {
	Authorization *authorization
	Categories    *categories
}

type Controller struct {
	Engine *gin.Engine
	combine
}

func AddController(engine *gin.Engine, repo *repository.Repository, strcode *strcode.Strcode, postman *mailagent.MailAgent, tokenMaker *authjwt.AuthJWT) *Controller {
	controllers := &Controller{
		Engine: engine,
		combine: combine{
			Categories:    newCategories(&repository.Repository{}),
			Authorization: newAuthorization(repo, strcode, postman, tokenMaker),
		},
	}

	controllers.init()
	return controllers
}

func (c *Controller) init() {
	router := c.Engine.Group("api/v1")

	//authorization
	{
		router.POST("/auth/signUp.Org", c.Authorization.SignUpOrg)
		router.POST("/auth/signIn.Org", c.Authorization.SignInOrg)
	}

	//categories
	{
		router.GET("/categories", c.Categories.GetAll)
	}
}
