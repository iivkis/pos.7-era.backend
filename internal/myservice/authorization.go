package myservice

import (
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
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
	Password string `json:"password" binding:"required,max=45"`
}

type signUpOrgOutput struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

type signInInput struct {
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45"`
}

//@Summary Регистрация организации, либо сотрудника
//@Description Метод позволяет зарегистрировать организацию или сотрудника данной огранизации.
//@Description Регистрация сотрудника возможна только с `jwt токеном` организации.
//@Param type query string false "`org`(default) or `employee`"
//@Param json body signUpOrgInput true "Объект для регитсрации огранизации. Обязательные поля:`email`, `password`"
//@Accept json
//@Produce json
//@Success 201 {object} signUpOrgOutput "Возвращаемый объкт при регистрации огранизации"
//@Failure 401 {object} myServiceError
//@Router /auth/signUp [post]
func (s *authorization) SignUp(c *gin.Context) {
	switch tp := c.DefaultQuery("type", "org"); tp {

	//case for organization
	case "org":
		var input signUpOrgInput

		//parse body
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusUnauthorized, errIncorrectInputData(err.Error()))
			return
		}

		//validate email
		if _, err := mail.ParseAddress(input.Email); err != nil {
			c.JSON(http.StatusUnauthorized, errIncorrectEmail(err.Error()))
			return
		}

		//create model and add in db
		model := repository.OrganizationModel{
			Name:     input.Name,
			Email:    input.Email,
			Password: input.Password,
		}

		if err := s.repo.Organizations.Create(&model); err != nil {
			c.String(http.StatusUnauthorized, err.Error())
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
		c.JSON(http.StatusUnauthorized, errIncorrectQueryType("incorrect argument for parametr `type`"))
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
