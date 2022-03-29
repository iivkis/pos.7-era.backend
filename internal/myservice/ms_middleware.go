package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
)

type MiddlewareService struct {
	repo    *repository.Repository
	authjwt *authjwt.AuthJWT
}

func newMiddlewareService(repo *repository.Repository, authjwt *authjwt.AuthJWT) *MiddlewareService {
	return &MiddlewareService{
		repo:    repo,
		authjwt: authjwt,
	}
}

func (s *MiddlewareService) AuthOrg() func(*gin.Context) {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			NewResponse(c, http.StatusUnauthorized, errUndefinedJWT())
			c.Abort()
			return
		}

		claims, err := s.authjwt.ParseOrganizationToken(token)
		if err != nil {
			NewResponse(c, http.StatusUnauthorized, errParsingJWT(err.Error()))
			c.Abort()
			return
		}

		c.Set("claims", claims)
	}
}

func (s *MiddlewareService) AuthEmployee(allowedRoles ...string) func(*gin.Context) {
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
		claims, err := s.authjwt.ParseEmployeeToken(token)
		if err != nil {
			NewResponse(c, http.StatusUnauthorized, errParsingJWT(err.Error()))
			c.Abort()
			return
		}

		//проверка прав доступа
		if !isAllowed(claims.Role) {
			NewResponse(c, http.StatusUnauthorized, errNoAccessRights())
			c.Abort()
			return
		}

		c.Set("claims", claims)
	}
}

type MiddlewareStdQueryInput struct {
	OutletID uint `form:"outlet_id"`
	OrgID    uint `form:"org_id"`

	Offset int `form:"offset"`
	Limit  int `form:"limit"`
}

//Standart Query
func (s *MiddlewareService) StdQuery() func(*gin.Context) {
	return func(c *gin.Context) {
		var query MiddlewareStdQueryInput
		if err := c.ShouldBindQuery(&query); err != nil {
			NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
			c.Abort()
			return
		}
		c.Set("std_query", &query)
	}
}
