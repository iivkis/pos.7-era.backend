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
	Employees     *employees
	Middleware    *middleware
}

type Controller struct {
	Engine *gin.Engine
	combine
}

func AddController(engine *gin.Engine, repo *repository.Repository, strcode *strcode.Strcode, postman *mailagent.MailAgent, tokenMaker *authjwt.AuthJWT) *Controller {
	controllers := &Controller{
		Engine: engine,
		combine: combine{
			Authorization: newAuthorization(repo, strcode, postman, tokenMaker),
			Categories:    newCategories(repo),
			Employees:     newEmployees(repo),
			Middleware:    newMiddleware(repo, tokenMaker),
		},
	}

	controllers.init()
	return controllers
}

func (c *Controller) init() {
	r := c.Engine.Group("api/v1")
	r.Use(c.Middleware.StdQuery())

	//authorization
	{
		//регистрация организации и сотрудника
		r.POST("/auth/signUp.Org", c.Authorization.SignUpOrg)
		r.POST("/auth/signUp.Employee", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Authorization.SignUpEmployee)

		//вход в аккаунт организации и сотрудника
		r.POST("/auth/signIn.Org", c.Authorization.SignInOrg)
		r.POST("/auth/signIn.Employee", c.Middleware.AuthOrg(), c.Authorization.SignInEmployee)

		//отправка код подтверждения на email и проверка
		r.GET("/auth/sendCode", c.Middleware.AuthOrg(), c.Authorization.SendCode)
		r.GET("/auth/confirmCode", c.Authorization.ConfirmCode)
	}

	//api для сотрудников
	{
		r.GET("/employees", c.Middleware.AuthOrg(), c.Employees.GetAll)
		r.PUT("/employees/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Employees.Update)
		r.DELETE("/employees/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Employees.Delete)
	}

	//categories
	{
		r.GET("/categories", c.Categories.GetAll)
	}
}
