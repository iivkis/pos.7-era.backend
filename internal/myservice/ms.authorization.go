package myservice

import (
	"errors"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"gorm.io/gorm"
)

type AuthorizationService interface {
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
	EmailConfirm(c *gin.Context)
}

type authorization struct {
	repo repository.Repository
}

func newAuthorizationService(repo repository.Repository) *authorization {
	return &authorization{
		repo: repo,
	}
}

type signUpOrgInput struct {
	Name     string `json:"name" binding:"required,max=45"`
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45,min=6"`
}

type signInOrgInput struct {
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45"`
}

type signInOrgOutput struct {
	Token string `json:"token"`
}

//@Summary Регистрация организации, либо сотрудника
//@Description Метод позволяет зарегистрировать организацию или сотрудника данной огранизации.
//@Description Регистрация сотрудника возможна только с `jwt токеном` организации.
//@Param type query string false "`org`(default) or `employee`"
//@Param json body signUpOrgInput true "Объект для регитсрации огранизации. Обязательные поля:`email`, `password`"
//@Accept json
//@Produce json
//@Success 201 {object} object "Возвращаемый объект при регистрации огранизации"
//@Failure 401 {object} serviceError
//@Router /auth/signUp [post]
func (s *authorization) SignUp(c *gin.Context) {
	switch tp := c.DefaultQuery("type", "org"); tp {
	//case for organization
	case "org":
		//parse body
		var input signUpOrgInput
		if err := c.ShouldBindJSON(&input); err != nil {
			newResponse(c, http.StatusUnauthorized, errIncorrectInputData(err.Error()))
			return
		}

		//validate email
		if _, err := mail.ParseAddress(input.Email); err != nil {
			newResponse(c, http.StatusUnauthorized, errIncorrectEmail(err.Error()))
			return
		}

		//create model and add in db
		model := repository.OrganizationModel{
			Name:     input.Name,
			Email:    input.Email,
			Password: input.Password,
		}

		if err := s.repo.Organizations.Create(&model); err != nil {
			if dberr, ok := isDatabaseError(err); ok {
				switch dberr.Number {
				case 1062:
					newResponse(c, http.StatusUnauthorized, errEmailExists())
				default:
					newResponse(c, http.StatusUnauthorized, errUnknownDatabase(dberr.Error()))
				}
				return
			}
			newResponse(c, http.StatusUnauthorized, errUnknownServer(err.Error()))
			return
		}

		newResponse(c, http.StatusCreated, nil)
	//case for employee
	case "employee":

	//default case
	default:
		newResponse(c, http.StatusUnauthorized, errIncorrectQuery())
	}
}

//@Summary Вход для организации, либо сотрудника
//@Description Метод позволяет войти в аккаунт организации, либо сотрудника.
//@Description Для входа сотрудника требуется `jwt токен` соотвествующей ему огранизации.
//@Param type query string false "`org`(default) or `employee`"
//@Param json body signInOrgInput true "Объект для входа в огранизацию. Обязательные поля:`email`, `password`"
//@Accept json
//@Produce json
//@Success 200 {object} signInOrgOutput "Возвращает `jwt токен` при успешной авторизации"
//@Failure 401 {object} serviceError
//@Router /auth/signIn [post]
func (s *authorization) SignIn(c *gin.Context) {
	switch c.DefaultQuery("type", "org") {
	case "org":
		var input signInOrgInput
		if err := c.ShouldBindJSON(&input); err != nil {
			newResponse(c, http.StatusUnauthorized, errIncorrectInputData(err.Error()))
			return
		}

		token, err := s.repo.Organizations.SignIn(input.Email, input.Password)
		if err != nil {
			if dberr, ok := isDatabaseError(err); ok {
				newResponse(c, http.StatusUnauthorized, errUnknownDatabase(dberr.Error()))
				return
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newResponse(c, http.StatusUnauthorized, errEmailNotFound())
				return
			}
			newResponse(c, http.StatusUnauthorized, errUnknownServer(err.Error()))
			return
		}

		output := signInOrgOutput{Token: token}
		newResponse(c, http.StatusOK, output)
	case "employee":
	default:
		newResponse(c, http.StatusUnauthorized, errIncorrectQuery("unknown argument for parametr `type`"))
	}
}

//@Summary Подтверждение email адреса
//@Param type query string false "`org`(default) or `employee`"
//@Param email query string false "параметр `email` хранит в себе почтовый адрес получателя письма с кодом для подтверждения"
//@Param code query string false "параметр `code` хранит в себе код подтверждения из письма. Данный параметр игнорируется при заданном параметре `email`"
//@Success 200 {object} object "Возвращает пустой объект в случае успеха"
//@Failure 400 {object} serviceError
//@Router /auth/emailConfirm [get]
func (s *authorization) EmailConfirm(c *gin.Context) {
	switch c.DefaultQuery("type", "org") {
	case "org":

	default:
		newResponse(c, http.StatusUnauthorized, errIncorrectQuery("unknown argument for parametr `type`"))
	}
}
