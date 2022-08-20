package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/tokenmaker"
)

func mustGetEmployeeClaims(c *gin.Context) *tokenmaker.EmployeeClaims {
	return c.MustGet("claims").(*tokenmaker.EmployeeClaims)
}

func mustGetOrganizationClaims(c *gin.Context) *tokenmaker.OrganizationClaims {
	return c.MustGet("claims").(*tokenmaker.OrganizationClaims)
}

func mustGetStandartQuery(c *gin.Context) *MiddlewareStandartQuery {
	return c.MustGet("std_query").(*MiddlewareStandartQuery)
}
