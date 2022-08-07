package controller

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/s3cloud"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"github.com/iivkis/pos.7-era.backend/pkg/mailagent"
	"github.com/iivkis/strcode"
)

type combine struct {
	Authorization            *authorization
	Outlets                  *outlets
	Categories               *categories
	Employees                *employees
	Products                 *products
	Ingredients              *ingredients
	IngredientsAddingHistory *ingredientsAddingHistory
	ProductsWithIngredients  *productsWithIngredients
	OrderInfo                *orderInfo
	OrderList                *orderList
	Orders                   *orders
	InventoryHistory         *inventoryHistory
	InventoryList            *inventoryList
	Invitation               *invitation
	Sessions                 *sessions
	Upload                   *upload
	CashChanges              *cashChanges
	Middleware               *middleware
}

type Controller struct {
	Engine *gin.Engine
	combine
}

func AddController(
	engine *gin.Engine,
	repo *repository.Repository,
	strcode *strcode.Strcode,
	postman *mailagent.MailAgent,
	tokenMaker *authjwt.AuthJWT,
	s3cloud *s3cloud.SelectelS3Cloud,
) *Controller {

	controllers := &Controller{
		Engine: engine,
		combine: combine{
			Authorization:            newAuthorization(repo, strcode, postman, tokenMaker),
			Outlets:                  newOutlets(repo),
			Categories:               newCategories(repo),
			Employees:                newEmployees(repo),
			Products:                 newProducts(repo, s3cloud),
			Ingredients:              newIngredients(repo),
			IngredientsAddingHistory: newIngredientsAddingHistory(repo),
			ProductsWithIngredients:  newProductsWithIngredients(repo),
			OrderInfo:                newOrderInfo(repo),
			OrderList:                newOrderList(repo),
			Orders:                   newOrders(repo),
			InventoryHistory:         newInventoryHistory(repo),
			InventoryList:            newInventoryList(repo),
			Invitation:               newInvitation(repo),
			Sessions:                 newSessions(repo),
			Upload:                   newUpload(repo, s3cloud),
			CashChanges:              newCashChanges(repo),
			Middleware:               newMiddleware(repo, tokenMaker),
		},
	}

	controllers.init()
	return controllers
}

func (c *Controller) init() {
	r := c.Engine.Group("api/v1")

	//middleware
	{
		r.Use(c.cors())
		r.Use(c.Middleware.StdQuery())
	}

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
		r.POST("/ingredients.Arrival", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.Ingredients.Arrival)

		// история добавления ингредиентов
		r.POST("/ingredients.History", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.IngredientsAddingHistory.Create)
		r.GET("/ingredients.History", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.IngredientsAddingHistory.GetAll)

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

	//чеки новое api
	{
		r.GET("/orders", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Orders.GetAll)
		r.POST("/orders", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Orders.Create)
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

	//списание определенных ингредиентов при покупке продуктов
	{
		r.GET("/pwis", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.ProductsWithIngredients.GetAll)
		r.POST("/pwis", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.ProductsWithIngredients.Create)
		r.PUT("/pwis/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.ProductsWithIngredients.Update)
		r.DELETE("/pwis/:id", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.ProductsWithIngredients.Delete)
	}

	//список изменения баланса кассы
	{
		r.GET("cashChanges", c.Middleware.AuthEmployee(r_owner, r_director), c.CashChanges.GetAll)
		r.GET("cashChanges.CurrentSession", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.CashChanges.GetAllForCurrentSession)
		r.POST("cashChanges", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.CashChanges.Create)
	}

	// история инвентпризации
	{
		r.GET("/inventoryHistory", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.InventoryHistory.GetAll)
		r.POST("/inventoryHistory", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.InventoryHistory.Create)
	}

	// список истории инвентаризации
	{
		r.GET("/inventoryList", c.Middleware.AuthEmployee(r_owner, r_director, r_admin), c.InventoryList.GetAll)
		r.POST("/inventoryList", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.InventoryList.Create)
	}

	// создание филиалов
	{
		r.POST("/invites", c.Middleware.AuthEmployee(r_owner, r_director), c.Invitation.Create)
		r.GET("/invites", c.Middleware.AuthEmployee(r_owner, r_director), c.Invitation.GetAll)
		r.GET("/invites.NotActivated", c.Middleware.AuthEmployee(r_owner, r_director), c.Invitation.GetNotActivated)
		r.GET("/invites.Activated", c.Middleware.AuthEmployee(r_owner, r_director), c.Invitation.GetActivated)
		r.DELETE("/invites/:id", c.Middleware.AuthEmployee(r_owner, r_director), c.Invitation.Delete)
	}

	{
		r.POST("/upload.Photo", c.Middleware.AuthEmployee(r_owner, r_director, r_admin, r_cashier), c.Upload.UploadPhoto)
	}
}

func (c *Controller) cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		MaxAge:           12 * time.Hour,
	})
}
