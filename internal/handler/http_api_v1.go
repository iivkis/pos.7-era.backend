package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

func (h *HttpHandler) connectApiV1(r *gin.RouterGroup) {
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "all okey!")
	})

	//authorization
	{
		//регистрация организации и сотрудника
		r.POST("/auth/signUp.Org", h.srv.Authorization.SignUpOrg)
		r.POST("/auth/signUp.Employee", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Authorization.SignUpEmployee)

		//вход в аккаунт организации и сотрудника
		r.POST("/auth/signIn.Org", h.srv.Authorization.SignInOrg)
		r.POST("/auth/signIn.Employee", h.srv.Mware.AuthOrg(), h.srv.Authorization.SignInEmployee)

		//отправка код подтверждения на email и проверка
		r.GET("/auth/sendCode", h.srv.Mware.AuthOrg(), h.srv.Authorization.SendCode)
		r.GET("/auth/confirmCode", h.srv.Authorization.ConfirmCode)
	}

	//api для сотрудников
	{
		r.GET("/employees", h.srv.Mware.AuthOrg(), h.srv.Employees.GetAll)
	}

	//api для торговых точек
	{
		r.POST("/outlets", h.srv.Mware.AuthOrg(), h.srv.Outlets.Create)
		r.GET("/outlets", h.srv.Mware.AuthOrg(), h.srv.Outlets.GetAll)
	}

	//api для сессий
	{
		r.POST("/sessions", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.Sessions.OpenOrClose)
		r.GET("/sessions", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Sessions.GetAll)

		r.GET("/sessions/last", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.Sessions.GetLastClosedForOutlet)
		r.GET("/sessions.Last", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.Sessions.GetLastForOutlet)
		r.GET("/sessions.Last.Closed", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.Sessions.GetLastClosedForOutlet)

	}

	//api для категорий
	{
		r.GET("/categories", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.Categories.GetAll)
		r.POST("/categories", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.Categories.Create)
		r.PUT("/categories/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.Categories.Update)
		r.DELETE("/categories/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.Categories.Delete)
	}

	//api для продуктов
	{
		r.GET("/products", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.Products.GetAllForOrg)
		r.GET("/products/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin, repository.R_CASHIER), h.srv.Products.GetOneForOutlet)
		r.POST("/products", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.Products.Create)
		r.PUT("/products/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.Products.UpdateFields)
		r.DELETE("/products/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.Products.Delete)

		r.GET("/products.Outlet", h.srv.Mware.AuthEmployee(r_cashier), h.srv.Products.GetAllForOutlet)
	}

	//ingredients api
	{
		r.POST("/ingredients", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Ingredients.Create)
		r.GET("/ingredients", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.Mware.StdQuery(), h.srv.Ingredients.GetAllForOrg)
		r.PUT("/ingredients/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.Ingredients.UpdateFields)
		r.DELETE("/ingredients/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.Ingredients.Delete)
	}

	//products with ingredients
	{
		r.GET("/pwis", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.ProductsWithIngredients.GetAllForOrg)
		r.POST("/pwis", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.ProductsWithIngredients.Create)
		r.PUT("/pwis/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.ProductsWithIngredients.UpdateFields)
		r.POST("/pwis/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin), h.srv.Mware.StdQuery(), h.srv.ProductsWithIngredients.Delete)
	}

	//order info
	{
		r.GET("/orderInfo", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.OrdersInfo.GetAllForOrg)
		r.POST("/orderInfo", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.OrdersInfo.Create)
		r.DELETE("/orderInfo/:id", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.OrdersInfo.Delete)
	}

	//order list
	{
		r.GET("/orderList", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.OrdersList.GetAllForOrg)
		r.POST("/orderList", h.srv.Mware.AuthEmployee(r_owner, r_admin, r_cashier), h.srv.OrdersList.Create)
	}
}
