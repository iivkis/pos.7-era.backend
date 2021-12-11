package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *HttpHandler) connectApiV1(r *gin.RouterGroup) {
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "all okey!")
	})

	//authorization
	authApi := r.Group("/auth")
	{
		authApi.POST("/signUp", h.service.Authorization.SignUp)
		authApi.POST("/signIn", h.service.Authorization.SignIn)
		authApi.GET("/sendCode", h.service.Authorization.SendCode)
		authApi.GET("/confirmCode", h.service.Authorization.ConfirmCode)
	}
}
