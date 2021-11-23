package myservice

import (
	"net/http"
	"net/mail"

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

type signUpInput struct {
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45"`
}

type signInInput struct {
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45"`
}

//@Summary Регистрация организации, либо сотрудника
//@Param type query string false "`org`(default) or `employee`"
//@Param json body signUpInput true "Объект с обязательными полями `email` и `password`"
//@Accept json
//@Produce plain
//@Success 200 {string} string "Возвращает `created` при успешной регистрации"
//@Failure 400 {string} string
//@Router /auth/signUp [post]
func (s *authorization) SignUp(c *gin.Context) {
	switch tp := c.DefaultQuery("type", "org"); tp {
	case "org":
		var input signUpInput

		//parse body
		if err := c.ShouldBindJSON(&input); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		//validate email
		if _, err := mail.ParseAddress(input.Email); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		model := repository.OrganizationModel{
			Email:    input.Email,
			Password: input.Password,
		}

		if err := s.repo.Organizations.Create(&model); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.String(http.StatusOK, "created")

	case "employee":
	default:
		c.String(http.StatusBadRequest, "undefined argument `%s` in parametr `type`", tp)
	}
}

//@Summary Вход для организации, либо сотрудника
//@Param type query string false "`org`(default) or `employee`"
//@Param json body signInInput true "Объект с обязательными полями `email` и `password`"
//@Accept json
//@Produce plain
//@Success 200 {string} string "Возвращает `jwt токен` при успешной авторизации"
//@Failure 400 {string} string
//@Router /auth/signIn [post]
func (s *authorization) SignIn(c *gin.Context) {
	switch tp := c.DefaultQuery("type", "org"); tp {
	case "org":
		var input signInInput

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
