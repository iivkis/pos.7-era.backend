package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/selectelS3Cloud"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"github.com/iivkis/pos.7-era.backend/pkg/mailagent"
	"github.com/iivkis/strcode"
)

type combine struct {
	Authorization *authorization
	Outlets       *outlets
	Categories    *categories
	Employees     *employees
	Products      *products
	Ingredients   *ingredients
	OrdersList    *orderList
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
			Outlets:       newOutlets(repo),
			Categories:    newCategories(repo),
			Employees:     newEmployees(repo),
			Products:      newProducts(repo, &selectelS3Cloud.SelectelS3Cloud{}),
			Ingredients:   newIngredients(repo),
			OrdersList:    newOrderList(repo),
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

	//api для торговых точек
	{
		r.GET("/outlets", c.Middleware.AuthOrg(), c.Outlets.GetAll)
		r.POST("/outlets", c.Middleware.AuthEmployee(r_owner, r_director), c.Outlets.Create)
		r.PUT("/outlets/:id", c.Middleware.AuthEmployee(r_owner, r_director), c.Outlets.Update)
		r.DELETE("/outlets/:id", c.Middleware.AuthEmployee(r_owner, r_director), c.Outlets.Delete)
	}

	//categories
	{
		r.GET("/categories", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Categories.GetAll)
		r.POST("/categories", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Categories.Create)
		r.PUT("/categories/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Categories.UpdateFields)
		r.DELETE("/categories/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Categories.Delete)
	}

	//ingredients api
	{
		r.POST("/ingredients", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Ingredients.Create)
		r.GET("/ingredients", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Ingredients.GetAll)
		r.PUT("/ingredients/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Ingredients.UpdateFields)
		r.DELETE("/ingredients/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Ingredients.Delete)

		//поступление ингредиентов
		// r.POST("/ingredients.Arrival", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Ingredients.Arrival)
	}

	//api для продуктов
	{
		r.GET("/products", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Products.GetAll)
		r.GET("/products/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, repository.R_CASHIER), c.Products.GetOne)
		r.POST("/products", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Products.Create)
		r.PUT("/products/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Products.Update)
		r.DELETE("/products/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Products.Delete)
	}

	{
		r.GET("/orderList", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.OrdersList.GetAll)
		r.GET("/orderList.Calc", c.Middleware.AuthEmployee(r_owner, r_director), c.OrdersList.Calc)
		r.POST("/orderList", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.OrdersList.Create)
	}

	//api для сессий
	{
		r.POST("/sessions", h.srv.Mware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), h.srv.Sessions.OpenOrClose)
		r.GET("/sessions", h.srv.Mware.AuthEmployee(r_owner, r_director, r_admin), h.srv.Sessions.GetAll)
		r.GET("/sessions.Last", h.srv.Mware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), h.srv.Sessions.GetLastForOutlet)
		r.GET("/sessions.Last.Me", h.srv.Mware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), h.srv.Sessions.GetLastForMe)
		r.GET("/sessions.Last.Closed", h.srv.Mware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), h.srv.Sessions.GetLastClosedForOutlet)
	}
}
