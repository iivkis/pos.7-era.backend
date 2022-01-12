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
		authApi.POST("/signUp.Org", h.service.Authorization.SignUpOrg)
		authApi.POST("/signUp.Employee", h.withAuthOrg(), h.service.Authorization.SignUpEmployee)

		authApi.POST("/signIn.Org", h.service.Authorization.SignInOrg)
		authApi.POST("/signIn.Employee", h.withAuthOrg(), h.service.Authorization.SignInEmployee)

		authApi.GET("/sendCode", h.service.Authorization.SendCode)
		authApi.GET("/confirmCode", h.service.Authorization.ConfirmCode)
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
