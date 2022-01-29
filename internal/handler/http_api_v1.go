package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/myservice"
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
		authApi.POST("/signUp.Employee", h.authEmployee("owner", "admin"), h.service.Authorization.SignUpEmployee)

		//вход в аккаунт организации и сотрудника
		authApi.POST("/signIn.Org", h.service.Authorization.SignInOrg)
		authApi.POST("/signIn.Employee", h.authOrg(), h.service.Authorization.SignInEmployee)

		//отправка код подтверждения на email и проверка
		authApi.GET("/sendCode", h.service.Authorization.SendCode)
		authApi.GET("/confirmCode", h.service.Authorization.ConfirmCode)
	}

	//api для сотрудников
	employeesApi := r.Group("/employees")
	{
		//список сотрудников организации
		employeesApi.GET("/", h.authOrg(), h.service.Employees.GetAll)
	}

	//api для торговых точек
	outletsApi := r.Group("/outlets")
	{
		//метод для добавления точки
		outletsApi.POST("/", h.authOrg(), h.service.Outlets.Create)
		outletsApi.GET("/", h.authOrg(), h.service.Outlets.GetAll)
	}

	sessionsApi := r.Group("/sessions")
	{
		sessionsApi.POST("/", h.authEmployee("owner", "admin", "cashier"), h.service.Session.OpenOrClose)
		sessionsApi.GET("/", h.authEmployee("owner", "admin"), h.service.Session.GetAll)
		sessionsApi.GET("/last", h.authEmployee("owner", "admin", "cashier"), h.service.Session.GetLastForOutlet)
	}

	categoryApi := r.Group("/category")
	{
		categoryApi.GET("/", h.authEmployee("owner", "admin", "cashier"), h.service.Category.GetAll)
		categoryApi.POST("/", h.authEmployee("owner", "admin"), h.service.Category.Create)
		categoryApi.PUT("/:id", h.authEmployee("owner", "admin"), h.service.Category.Update)
		categoryApi.DELETE("/:id", h.authEmployee("owner", "admin"), h.service.Category.Delete)
	}
}

func (h *HttpHandler) authOrg() gin.HandlerFunc {
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

func (h *HttpHandler) authEmployee(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			myservice.NewResponse(c, http.StatusUnauthorized, myservice.ErrUndefinedJWT())
			c.Abort()
			return
		}

		claims, err := h.authjwt.ParseEmployeeToken(token)

		if err != nil {
			myservice.NewResponse(c, http.StatusUnauthorized, myservice.ErrParsingJWT(err.Error()))
			c.Abort()
			return
		}

		//проверка прав доступа
		{
			var allowed bool
			for _, role := range roles {
				if role == claims.Role {
					allowed = true
					break
				}
			}

			if !allowed {
				myservice.NewResponse(c, http.StatusUnauthorized, myservice.ErrNoAccessRights())
				c.Abort()
				return
			}
		}

		c.Set("claims_org_id", claims.OrganizationID)
		c.Set("claims_employee_id", claims.EmployeeID)
		c.Set("claims_outlet_id", claims.OutletID)
	}
}
