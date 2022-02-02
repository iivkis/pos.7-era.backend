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
	authApi := r.Group("/auth")
	{
		//регистрация организации и сотрудника
		authApi.POST("/signUp.Org", h.service.Authorization.SignUpOrg)
		authApi.POST("/signUp.Employee", h.withAuthEmployee(repository.R_OWNER, repository.R_ADMIN), h.service.Authorization.SignUpEmployee)

		//вход в аккаунт организации и сотрудника
		authApi.POST("/signIn.Org", h.service.Authorization.SignInOrg)
		authApi.POST("/signIn.Employee", h.withAuthOrg(), h.service.Authorization.SignInEmployee)

		//отправка код подтверждения на email и проверка
		authApi.GET("/sendCode", h.withAuthOrg(), h.service.Authorization.SendCode)
		authApi.GET("/confirmCode", h.service.Authorization.ConfirmCode)
	}

	//api для сотрудников
	employeesApi := r.Group("/employees")
	{
		employeesApi.GET("/", h.withAuthOrg(), h.service.Employees.GetAll)
	}

	//api для торговых точек
	outletsApi := r.Group("/outlets")
	{
		outletsApi.POST("/", h.withAuthOrg(), h.service.Outlets.Create)
		outletsApi.GET("/", h.withAuthOrg(), h.service.Outlets.GetAll)
	}

	//api для сессий
	sessionsApi := r.Group("/sessions")
	{
		sessionsApi.POST("/", h.withAuthEmployee(repository.R_OWNER, repository.R_ADMIN, repository.R_CASHIER), h.service.Sessions.OpenOrClose)
		sessionsApi.GET("/", h.withAuthEmployee(repository.R_OWNER, repository.R_ADMIN), h.service.Sessions.GetAll)
		sessionsApi.GET("/last", h.withAuthEmployee(repository.R_OWNER, repository.R_ADMIN, repository.R_CASHIER), h.service.Sessions.GetLastForOutlet)
	}

	//api для категорий
	categoriesApi := r.Group("/categories")
	{
		categoriesApi.GET("/", h.withAuthEmployee(repository.R_OWNER, repository.R_ADMIN, repository.R_CASHIER), h.service.Categories.GetAll)
		categoriesApi.POST("/", h.withAuthEmployee(repository.R_OWNER, repository.R_ADMIN), h.service.Categories.Create)
		categoriesApi.PUT("/:id", h.withAuthEmployee(repository.R_OWNER, repository.R_ADMIN), h.service.Categories.Update)
		categoriesApi.DELETE("/:id", h.withAuthEmployee(repository.R_OWNER, repository.R_ADMIN), h.service.Categories.Delete)
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
