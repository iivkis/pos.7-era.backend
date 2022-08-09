package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/tokenmaker"
)

type middleware struct {
	repo       *repository.Repository
	tokenMaker *tokenmaker.TokenMaker
}

func newMiddleware(repo *repository.Repository, authjwt *tokenmaker.TokenMaker) *middleware {
	return &middleware{
		repo:       repo,
		tokenMaker: authjwt,
	}
}

func (s *middleware) AuthOrg() func(*gin.Context) {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			NewResponse(c, http.StatusUnauthorized, errUndefinedJWT())
			c.Abort()
			return
		}

		claims, err := s.tokenMaker.ParseOrganizationToken(token)
		if err != nil {
			NewResponse(c, http.StatusUnauthorized, errParsingJWT(err.Error()))
			c.Abort()
			return
		}

		c.Set("claims", claims)
	}
}

func (s *middleware) AuthEmployee(allowedRoles ...string) func(*gin.Context) {
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
			NewResponse(c, http.StatusUnauthorized, errUndefinedJWT())
			c.Abort()
			return
		}

		//парсинг токена
		claims, err := s.tokenMaker.ParseEmployeeToken(token)
		if err != nil {
			NewResponse(c, http.StatusUnauthorized, errParsingJWT(err.Error()))
			c.Abort()
			return
		}

		//проверка прав доступа
		if !isAllowed(claims.Role) {
			NewResponse(c, http.StatusUnauthorized, errPermissionDenided())
			c.Abort()
			return
		}

		c.Set("claims", claims)
	}
}

type MiddlewareStandartQuery struct {
	OutletID uint `form:"outlet_id" binding:"min=0"`
	OrgID    uint `form:"org_id" binding:"min=0"`

	Offset int `form:"offset" binding:"min=0"`
	Limit  int `form:"limit" binding:"min=0"`
}

// Standart Query
func (s *middleware) StandartQuery() func(*gin.Context) {
	return func(c *gin.Context) {
		var query MiddlewareStandartQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
			c.Abort()
			return
		}
		c.Set("std_query", &query)
	}
}
