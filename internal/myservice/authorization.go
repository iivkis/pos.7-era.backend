package myservice

import (
	"errors"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"gorm.io/gorm"
)

//BasePath /auth/

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

type signUpOrgInput struct {
	Name     string `json:"name" binding:"required,max=45"`
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45,min=6"`
}

type signUpOrgOutput struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
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
//@Success 201 {object} signUpOrgOutput "Возвращаемый объект при регистрации огранизации"
//@Failure 401 {object} myServiceError
//@Router /auth/signUp [post]
func (s *authorization) SignUp(c *gin.Context) {
	switch tp := c.DefaultQuery("type", "org"); tp {
	//case for organization
	case "org":
		//parse body
		var input signUpOrgInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusUnauthorized, ERR_INCORRECT_INPUT_DATA(err.Error()))
			return
		}

		//validate email
		if _, err := mail.ParseAddress(input.Email); err != nil {
			c.JSON(http.StatusUnauthorized, ERR_INCORRECT_EMAIL(err.Error()))
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
					c.JSON(http.StatusUnauthorized, ERR_EMAIL_ALREADY_EXISTS())
				default:
					c.JSON(http.StatusUnauthorized, ERR_UNKNOWN_DATABASE(dberr.Error()))
				}
				return
			}
			c.JSON(http.StatusUnauthorized, ERR_UNKNOWN_SERVER(err.Error()))
			return
		}

		//output result
		output := signUpOrgOutput{
			ID:    model.ID,
			Email: model.Email,
		}
		c.JSON(http.StatusCreated, output)

	//case for employee
	case "employee":

	//default case
	default:
		c.JSON(http.StatusUnauthorized, ERR_INCORRECT_QUERY_TYPE())
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
//@Failure 401 {object} myServiceError
//@Router /auth/signIn [post]
func (s *authorization) SignIn(c *gin.Context) {
	switch queryType := c.DefaultQuery("type", "org"); queryType {
	case "org":
		var input signInOrgInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusUnauthorized, ERR_INCORRECT_INPUT_DATA(err.Error()))
			return
		}

		token, err := s.repo.Organizations.SignIn(input.Email, input.Password)
		if err != nil {
			if dberr, ok := isDatabaseError(err); ok {
				c.JSON(http.StatusUnauthorized, ERR_UNKNOWN_DATABASE(dberr.Error()))
				return
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, ERR_EMAIL_NOT_FOUND())
				return
			}
			c.JSON(http.StatusUnauthorized, ERR_UNKNOWN_SERVER(err.Error()))
			return
		}

		output := signInOrgOutput{Token: token}
		c.JSON(http.StatusOK, output)
	case "employee":
	default:
		c.JSON(http.StatusUnauthorized, ERR_INCORRECT_QUERY_TYPE())
	}
}
