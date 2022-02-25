package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/myservice"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

func (h *HttpHandler) connectApiV1(r *gin.RouterGroup) {
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "all okey!")
	})

	//authorization
	{
		//регистрация организации и сотрудника
		r.POST("/auth/signUp.Org", h.service.Authorization.SignUpOrg)
		r.POST("/auth/signUp.Employee", h.withAuthEmployee(r_owner, r_admin), h.service.Authorization.SignUpEmployee)

		//вход в аккаунт организации и сотрудника
		r.POST("/auth/signIn.Org", h.service.Authorization.SignInOrg)
		r.POST("/auth/signIn.Employee", h.withAuthOrg(), h.service.Authorization.SignInEmployee)

		//отправка код подтверждения на email и проверка
		r.GET("/auth/sendCode", h.withAuthOrg(), h.service.Authorization.SendCode)
		r.GET("/auth/confirmCode", h.service.Authorization.ConfirmCode)
	}

	//api для сотрудников
	{
		r.GET("/employees", h.withAuthOrg(), h.service.Employees.GetAll)
	}

	//api для торговых точек
	{
		r.POST("/outlets", h.withAuthOrg(), h.service.Outlets.Create)
		r.GET("/outlets", h.withAuthOrg(), h.service.Outlets.GetAll)
	}

	//api для сессий
	{
		r.POST("/sessions", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.Sessions.OpenOrClose)
		r.GET("/sessions", h.withAuthEmployee(r_owner, r_admin), h.service.Sessions.GetAll)
		r.GET("/sessions/last", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.Sessions.GetLastForOutlet)
	}

	//api для категорий
	{
		r.GET("/categories", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.Categories.GetAll)
		r.POST("/categories", h.withAuthEmployee(r_owner, r_admin), h.service.Categories.Create)
		r.PUT("/categories/:id", h.withAuthEmployee(r_owner, r_admin), h.service.Categories.Update)
		r.DELETE("/categories/:id", h.withAuthEmployee(r_owner, r_admin), h.service.Categories.Delete)
	}

	//api для продуктов
	{
		r.GET("/products", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.Products.GetAllForOrg)
		r.GET("/products/:id", h.withAuthEmployee(r_owner, r_admin, repository.R_CASHIER), h.service.Products.GetOneForOutlet)
		r.POST("/products", h.withAuthEmployee(r_owner, r_admin), h.service.Products.Create)
		r.PUT("/products/:id", h.withAuthEmployee(r_owner, r_admin), h.service.Products.UpdateFields)
		r.DELETE("/products/:id", h.withAuthEmployee(r_owner, r_admin), h.service.Products.Delete)

		r.GET("/products.Outlet", h.withAuthEmployee(r_cashier), h.service.Products.GetAllForOutlet)
	}

	//ingredients api
	{
		r.POST("/ingredients", h.withAuthEmployee(r_owner, r_admin), h.service.Ingredients.Create)
		r.GET("/ingredients", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.Ingredients.GetAllForOrg)
		r.PUT("/ingredients/:id", h.withAuthEmployee(r_owner, r_admin), h.service.Ingredients.UpdateFields)
		r.DELETE("/ingredients/:id", h.withAuthEmployee(r_owner, r_admin), h.service.Ingredients.Delete)
	}

	//products with ingredients
	{
		r.GET("/pwis", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.ProductsWithIngredients.GetAllForOrg)
		r.POST("/pwis", h.withAuthEmployee(r_owner, r_admin), h.service.ProductsWithIngredients.Create)
		r.PUT("/pwis/:id", h.withAuthEmployee(r_owner, r_admin), h.service.ProductsWithIngredients.UpdateFields)
		r.POST("/pwis/:id", h.withAuthEmployee(r_owner, r_admin), h.service.ProductsWithIngredients.Delete)
	}

	//order info
	{
		r.GET("/orderInfo", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.OrdersInfo.GetAllForOrg)
		r.POST("/orderInfo", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.OrdersInfo.Create)
		r.DELETE("/orderInfo/:id", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.OrdersInfo.Delete)
	}

	//order list
	{
		r.GET("/orderList", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.OrdersList.GetAllForOrg)
		r.POST("/orderList", h.withAuthEmployee(r_owner, r_admin, r_cashier), h.service.OrdersList.Create)
	}

}

func (h *HttpHandler) withAuthOrg() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			myservice.NewResponse(c, http.StatusUnauthorized, myservice.ErrUndefinedJWT())
			c.Abort()
			return
		}

		claims, err := h.authjwt.ParseOrganizationToken(token)
		if err != nil {
			myservice.NewResponse(c, http.StatusUnauthorized, myservice.ErrParsingJWT(err.Error()))
			c.Abort()
			return
		}

		c.Set("claims_org_id", claims.OrganizationID)
	}
}

func (h *HttpHandler) withAuthEmployee(allowedRoles ...string) gin.HandlerFunc {
	//создаем карту с ролями для быстрого поиска
	var allowed = map[string]uint8{}
	for i, roles := range allowedRoles {
		allowed[roles] = uint8(i)
	}

	var isAllowed = func(role string) bool {
		_, ok := allowed[role]
		return ok
	}

	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			myservice.NewResponse(c, http.StatusUnauthorized, myservice.ErrUndefinedJWT())
			c.Abort()
			return
		}

		//парсинг токена
		claims, err := h.authjwt.ParseEmployeeToken(token)
		if err != nil {
			myservice.NewResponse(c, http.StatusUnauthorized, myservice.ErrParsingJWT(err.Error()))
			c.Abort()
			return
		}

		//проверка прав доступа
		if !isAllowed(claims.Role) {
			myservice.NewResponse(c, http.StatusUnauthorized, myservice.ErrNoAccessRights())
			c.Abort()
			return
		}

		c.Set("claims_org_id", claims.OrganizationID)
		c.Set("claims_outlet_id", claims.OutletID)
		c.Set("claims_employee_id", claims.EmployeeID)
	}
}
