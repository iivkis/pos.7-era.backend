package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

type AuthorizationService interface {
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
}

type authorization struct {
	repo repository.Repository
}

func newAuthorizationService(repo repository.Repository) *authorization {
	return &authorization{
		repo: repo,
	}
}

func (s *authorization) SignUp(c *gin.Context) {
	switch tp := c.DefaultQuery("type", "org"); tp {
	case "org":
		var input struct {
			Name     string `json:"name" binding:"max=45"`
			Email    string `json:"email" binding:"required,max=45"`
			Password string `json:"password" binding:"required,max=45"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		model := repository.OrganizationModel{
			Name:     input.Name,
			Email:    input.Email,
			Password: input.Password,
		}

		if err := s.repo.Organizations.Create(&model); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.String(http.StatusOK, "")

	case "employee":
	default:
		c.String(http.StatusBadRequest, "undefined argument `%s` in parametr `type`", tp)
	}
}

func (s *authorization) SignIn(c *gin.Context) {
	switch tp := c.DefaultQuery("type", "org"); tp {
	case "org":
		var input struct {
			Email    string `json:"email" binding:"required,max=45"`
			Password string `json:"password" binding:"required,max=45"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		token, err := s.repo.Organizations.SignIn(input.Email, input.Password)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.String(http.StatusOK, token)

	case "employee":
	default:
		c.String(http.StatusBadRequest, "undefined argument `%s` in parametr `type`", tp)
	}
}
