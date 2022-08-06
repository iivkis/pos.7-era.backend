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
	OrderInfo     *orderInfo
	OrderList     *orderList
	Sessions      *sessions

	Middleware *middleware
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
			OrderInfo:     newOrderInfo(repo),
			OrderList:     newOrderList(repo),
			Sessions:      newSessions(repo),
			Middleware:    newMiddleware(repo, tokenMaker),
		},
	}

	controllers.init()
	return controllers
}

func (c *Controller) init() {
	r := c.Engine.Group("api/v1")
	r.Use(c.Middleware.StdQuery())

	//авторизация
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

	//сотудники
	{
		r.GET("/employees", c.Middleware.AuthOrg(), c.Employees.GetAll)
		r.PUT("/employees/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Employees.Update)
		r.DELETE("/employees/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Employees.Delete)
	}

	//торговые точки
	{
		r.GET("/outlets", c.Middleware.AuthOrg(), c.Outlets.GetAll)
		r.POST("/outlets", c.Middleware.AuthEmployee(r_owner, r_director), c.Outlets.Create)
		r.PUT("/outlets/:id", c.Middleware.AuthEmployee(r_owner, r_director), c.Outlets.Update)
		r.DELETE("/outlets/:id", c.Middleware.AuthEmployee(r_owner, r_director), c.Outlets.Delete)
	}

	//категории
	{
		r.GET("/categories", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Categories.GetAll)
		r.POST("/categories", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Categories.Create)
		r.PUT("/categories/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Categories.UpdateFields)
		r.DELETE("/categories/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Categories.Delete)
	}

	//ингредиенты к продуктам
	{
		r.POST("/ingredients", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Ingredients.Create)
		r.GET("/ingredients", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Ingredients.GetAll)
		r.PUT("/ingredients/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Ingredients.UpdateFields)
		r.DELETE("/ingredients/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Ingredients.Delete)

		//поступление ингредиентов
		// r.POST("/ingredients.Arrival", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Ingredients.Arrival)
	}

	//проудкты
	{
		r.GET("/products", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Products.GetAll)
		r.GET("/products/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, repository.R_CASHIER), c.Products.GetOne)
		r.POST("/products", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Products.Create)
		r.PUT("/products/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Products.Update)
		r.DELETE("/products/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Products.Delete)
	}

	//чеки
	{
		r.GET("/orderInfo", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.OrderInfo.GetAll)
		r.POST("/orderInfo", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.OrderInfo.Create)
		r.DELETE("/orderInfo/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.OrderInfo.Delete)
		r.POST("/orderInfo/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.OrderInfo.Recovery)
	}

	//список купленных продуктов в чеке
	{
		r.GET("/orderList", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.OrderList.GetAll)
		r.GET("/orderList.Calc", c.Middleware.AuthEmployee(r_owner, r_director), c.OrderList.Calc)
		r.POST("/orderList", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.OrderList.Create)
	}

	//смены сотрудников
	{
		r.POST("/sessions", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Sessions.OpenOrClose)
		r.GET("/sessions", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Sessions.GetAll)
		r.GET("/sessions.Last", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Sessions.GetLastForOutlet)
		r.GET("/sessions.Last.Me", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Sessions.GetLastForMe)
		r.GET("/sessions.Last.Closed", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Sessions.GetLastClosedForOutlet)
	}
}
