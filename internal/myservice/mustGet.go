package myservice

import (
	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
)

func mustGetEmployeeClaims(c *gin.Context) *authjwt.EmployeeClaims {
	return c.MustGet("claims").(*authjwt.EmployeeClaims)
}

func mustGetOrganizationClaims(c *gin.Context) *authjwt.OrganizationClaims {
	return c.MustGet("claims").(*authjwt.OrganizationClaims)
}

func mustGetStdQuery(c *gin.Context) *MiddlewareStdQueryInput {
	return c.MustGet("std_query").(*MiddlewareStdQueryInput)
}
