package myservice

import (
	"errors"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/config"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"github.com/iivkis/pos-ninja-backend/pkg/mailagent"
	"github.com/iivkis/strcode"
	"gorm.io/gorm"
)

type AuthorizationService interface {
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
	SendCode(c *gin.Context)
	ConfirmCode(c *gin.Context)
}

type authorization struct {
	repo      repository.Repository
	strcode   *strcode.Strcode
	mailagent *mailagent.MailAgent
}

func newAuthorizationService(repo repository.Repository, strcode *strcode.Strcode, mailagent *mailagent.MailAgent) *authorization {
	return &authorization{
		repo:      repo,
		strcode:   strcode,
		mailagent: mailagent,
	}
}

type signUpOrgInput struct {
	Name     string `json:"name" binding:"required,max=45"`
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45,min=6"`
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
		{
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
		}
	//case for employee
	case "employee":
		{
		}
	//default case
	default:
		newResponse(c, http.StatusUnauthorized, errIncorrectInputData())
	}
}

type signInOrgInput struct {
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45"`
}

type signInOrgOutput struct {
	Token string `json:"token"`
}

//@Summary Вход для организации, либо сотрудника
//@Description Метод позволяет войти в аккаунт организации, либо сотрудника.
//@Description Для входа сотрудника требуется `jwt токен` соотвествующей ему огранизации.
//@Param type query string false "`org` or `employee`"
//@Param json body signInOrgInput true "Объект для входа в огранизацию. Обязательные поля:`email`, `password`"
//@Accept json
//@Produce json
//@Success 200 {object} signInOrgOutput "Возвращает `jwt токен` при успешной авторизации"
//@Failure 401 {object} serviceError
//@Router /auth/signIn [post]
func (s *authorization) SignIn(c *gin.Context) {
	switch c.Query("type") {
	case "org":
		{
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
		}
	case "employee":
		{
		}
	default:
		newResponse(c, http.StatusUnauthorized, errIncorrectInputData("unknown argument for parametr `type`"))
	}
}

type sendCodeInputQuery struct {
	Type  string `form:"type,min=1"`
	Email string `form:"email,min=1,max=45"`
}

//@Summary Отправка кода подтверждения почты
//@param type query string false "`org` or `employee`"
//@param email query string false "адрес на который будет отправлено письмо (например: email@exmp.ru)"
//@Success 200 {object} object "возвращает пустой объект"
//@Router /auth/sendCode [get]
func (s *authorization) SendCode(c *gin.Context) {
	var inputQ sendCodeInputQuery
	if err := c.ShouldBindQuery(&inputQ); err != nil {
		newResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	switch inputQ.Type {
	case "org":
		{
			if ok, err := s.repo.Organizations.EmailExists(inputQ.Email); err != nil {
				newResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
			} else if !ok {
				newResponse(c, http.StatusBadRequest, errEmailNotFound())
			} else {
				s.mailagent.SendTemplate(inputQ.Email, "confirm_code.html", mailagent.Value{
					"code":     s.strcode.Encode(inputQ.Email),
					"host":     config.Env.Host,
					"port":     config.Env.Port,
					"protocol": config.Env.Protocol,
					"type":     "org",
				})
				if err != nil {
					newResponse(c, http.StatusInternalServerError, errUnknownServer("error on send email"))
					return
				}
				newResponse(c, http.StatusOK, nil)
			}
		}
	case "employee":
		{

		}
	default:
		newResponse(c, http.StatusBadRequest, errIncorrectInputData())
	}

}

type confirmCodeInputQuery struct {
	Type string `form:"type,min=1"`
	Code string `form:"code,min=1"`
}

//@Summary Проверка кода подтверждения
//@param type query string false "`org` or `employee`"
//@param code query string false "адрес на который будет отправлено письмо (например: email@exmp.ru)"
//@Success 200 {object} object "возвращает пустой объект"
//@Router /auth/confirmCode [get]
func (s *authorization) ConfirmCode(c *gin.Context) {
	var inputQ confirmCodeInputQuery
	if err := c.ShouldBindQuery(&inputQ); err != nil {
		newResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	switch inputQ.Type {
	case "org":
		{
			email, err := s.strcode.Decode(inputQ.Code)
			if err != nil {
				newResponse(c, http.StatusBadRequest, errIncorrectConfirmCode(err.Error()))
				return
			}

			if err := s.repo.Organizations.ConfirmEmailTrue(email); err != nil {
				newResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
				return
			}
		}
	default:
		newResponse(c, http.StatusBadRequest, errIncorrectInputData())
	}

	newResponse(c, http.StatusOK, nil)
}
