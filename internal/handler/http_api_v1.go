package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *HttpHandler) connectApiV1(r *gin.RouterGroup) {
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "all okey!")
	})

	//organizations
	orgApi := r.Group("/organizations")
	{
		orgApi.GET("/", h.service.Organizations.GetAllOrganizations)
		orgApi.POST("/", h.service.Organizations.AddOrganization)
	}
}
