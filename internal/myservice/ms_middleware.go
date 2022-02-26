package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
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
			NewResponse(c, http.StatusUnauthorized, ErrUndefinedJWT())
			c.Abort()
			return
		}

		claims, err := s.authjwt.ParseOrganizationToken(token)
		if err != nil {
			NewResponse(c, http.StatusUnauthorized, ErrParsingJWT(err.Error()))
			c.Abort()
			return
		}

		c.Set("claims_org_id", claims.OrganizationID)
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
			NewResponse(c, http.StatusUnauthorized, ErrUndefinedJWT())
			c.Abort()
			return
		}

		//парсинг токена
		claims, err := s.authjwt.ParseEmployeeToken(token)
		if err != nil {
			NewResponse(c, http.StatusUnauthorized, ErrParsingJWT(err.Error()))
			c.Abort()
			return
		}

		//проверка прав доступа
		if !isAllowed(claims.Role) {
			NewResponse(c, http.StatusUnauthorized, ErrNoAccessRights())
			c.Abort()
			return
		}

		c.Set("claims_org_id", claims.OrganizationID)
		c.Set("claims_outlet_id", claims.OutletID)
		c.Set("claims_employee_id", claims.EmployeeID)
		c.Set("claims_role", claims.Role)
	}
}

func (s *MiddlewareService) StdQuery() func(*gin.Context) {
	return func(c *gin.Context) {
		var query struct {
			OutletID uint `form:"outlet_id"`
		}

		if err := c.ShouldBindQuery(&query); err != nil {
			NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
			c.Abort()
			return
		}

		if query.OutletID != 0 &&
			c.MustGet("claims_role").(string) == repository.R_OWNER {
			if s.repo.Outlets.ExistsInOrg(query.OutletID, c.MustGet("claims_org_id")) {
				c.Set("claims_outlet_id", query.OutletID)
			} else {
				NewResponse(c, http.StatusBadRequest, errRecordNotFound("outlet with this ID undefined"))
				c.Abort()
				return
			}
		}
	}
}
