package myservice

import (
	"errors"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/config"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"github.com/iivkis/pos-ninja-backend/pkg/mailagent"
	"github.com/iivkis/strcode"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthorizationService struct {
	repo      repository.Repository
	strcode   *strcode.Strcode
	mailagent *mailagent.MailAgent
	authjwt   *authjwt.AuthJWT
}

func newAuthorizationService(repo repository.Repository, strcode *strcode.Strcode, mailagent *mailagent.MailAgent, authjwt *authjwt.AuthJWT) *AuthorizationService {
	return &AuthorizationService{
		repo:      repo,
		strcode:   strcode,
		mailagent: mailagent,
		authjwt:   authjwt,
	}
}

type signUpOrgInput struct {
	Name     string `json:"name" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=45"`
}

//@Summary Регистрация организации
//@Description Метод позволяет зарегистрировать организацию
//@Param json body signUpOrgInput true "Объект для регитсрации огранизации."
//@Accept json
//@Produce json
//@Success 201 {object} object "Возвращаемый объект при регистрации огранизации"
//@Failure 401 {object} serviceError
//@Router /auth/signUp.Org [post]
func (s *AuthorizationService) SignUpOrg(c *gin.Context) {
	//parse body
	var input signUpOrgInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	//validate email
	if _, err := mail.ParseAddress(input.Email); err != nil {
		NewResponse(c, http.StatusUnauthorized, errIncorrectEmail(err.Error()))
		return
	}

	//create orgModel and add in db
	orgModel := repository.OrganizationModel{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	}

	if err := s.repo.Organizations.Create(&orgModel); err != nil {
		if dberr, ok := isDatabaseError(err); ok {
			switch dberr.Number {
			case 1062:
				NewResponse(c, http.StatusUnauthorized, errEmailExists())
			default:
				NewResponse(c, http.StatusUnauthorized, errUnknownDatabase(dberr.Error()))
			}
			return
		}
		NewResponse(c, http.StatusUnauthorized, errUnknownServer(err.Error()))
		return
	}

	//Создание главной торговой точки
	outletModel := repository.OutletModel{
		Name:  "Главная точка продаж",
		OrgID: orgModel.ID,
	}

	if err := s.repo.Outlets.Create(&outletModel); err != nil {
		if dberr, ok := isDatabaseError(err); ok {
			NewResponse(c, http.StatusUnauthorized, errUnknownDatabase(dberr.Error()))
			return
		}
		NewResponse(c, http.StatusUnauthorized, errUnknownServer(err.Error()))
		return
	}

	//Создание аккаунта владельца
	employeeModelOwner := repository.EmployeeModel{
		Name:     "Управление организацией",
		Password: "000000",
		Role:     repository.R_OWNER,
		OrgID:    orgModel.ID,
		OutletID: outletModel.ID,
	}

	if err := s.repo.Employees.Create(&employeeModelOwner); err != nil {
		NewResponse(c, http.StatusUnauthorized, errUnknownDatabase(err.Error()))
		return
	}

	//Создание аккаунта кассира
	employeeModelCashier := repository.EmployeeModel{
		Name:     "Кассир (продажа товара)",
		Password: "000000",
		Role:     repository.R_CASHIER,
		OrgID:    orgModel.ID,
		OutletID: outletModel.ID,
	}

	if err := s.repo.Employees.Create(&employeeModelCashier); err != nil {
		NewResponse(c, http.StatusUnauthorized, errUnknownServer(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, nil)
}

type signUpEmployeeInput struct {
	Name     string `json:"name" binding:"required,min=2,max=200"`
	Password string `json:"password" binding:"required,min=6,max=6"`
	Role     string `json:"role" binding:"required,max=20"`
	OutletID uint   `json:"outlet_id" binding:"min=1"`
}

//@Summary Регистрация сотрудника
//@Description Метод позволяет зарегистрировать ссотрудника. Работает только с токеном организации.
//@Param json body signUpEmployeeInput true "Объект для регитсрации сотрудника."
//@Accept json
//@Produce json
//@Success 201 {object} object "Возвращаемый объект при регистрации сотрудника"
//@Failure 400 {object} serviceError
//@Router /auth/signUp.Employee [post]
func (s *AuthorizationService) SignUpEmployee(c *gin.Context) {
	var orgID = c.MustGet("claims_org_id").(uint)

	//parse JSON body
	var input signUpEmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	if input.Role == repository.R_OWNER {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("you cannot create a user with the owner role"))
		return
	}

	//create model and add
	model := repository.EmployeeModel{
		Name:     input.Name,
		Password: input.Password,
		Role:     input.Role,
		OutletID: input.OutletID,
		OrgID:    orgID,
	}

	if err := s.repo.Employees.Create(&model); err != nil {
		if errors.Is(err, repository.ErrOnlyNumInPassword) || errors.Is(err, repository.ErrUndefinedRole) {
			NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
			return
		}
		NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusCreated, nil)
}

type signInOrgInput struct {
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45"`
}

type signInOrgOutput struct {
	Token string `json:"token"`
}

//@Summary Вход для организации
//@Description Метод позволяет войти в аккаунт организации.
//@Param json body signInOrgInput true "Объект для входа в огранизацию."
//@Accept json
//@Produce json
//@Success 200 {object} signInOrgOutput "Возвращает `jwt токен` при успешной авторизации"
//@Failure 401 {object} serviceError
//@Router /auth/signIn.Org [post]
func (s *AuthorizationService) SignInOrg(c *gin.Context) {
	var input signInOrgInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusUnauthorized, errIncorrectInputData(err.Error()))
		return
	}

	org, err := s.repo.Organizations.SignIn(input.Email, input.Password)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusUnauthorized, errEmailNotFound())
			return
		}
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			NewResponse(c, http.StatusUnauthorized, errIncorrectPassword())
			return
		}
		NewResponse(c, http.StatusUnauthorized, errUnknownDatabase(err.Error()))
		return
	}

	claims := authjwt.OrganizationClaims{
		OrganizationID: org.ID,
	}

	token, err := s.authjwt.SignInOrganization(&claims)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownServer(err.Error()))
		return
	}

	output := signInOrgOutput{Token: token}
	NewResponse(c, http.StatusOK, output)
}

type signInEmployeeInput struct {
	ID       uint   `json:"id"`
	Password string `json:"password" binding:"required,max=45"`
}

type signInEmployeeOutput struct {
	Token string `json:"token"`
}

//@Summary Вход для сотрудника
//@Description Метод позволяет войти в аккаунт сотрудника. Работает только с токеном огранизации.
//@Param json body signInEmployeeInput true "Объект для входа в огранизацию."
//@Accept json
//@Produce json
//@Success 200 {object} signInEmployeeOutput "Возвращает `jwt токен` при успешной авторизации"
//@Failure 401 {object} serviceError
//@Router /auth/signIn.Employee [post]
func (s *AuthorizationService) SignInEmployee(c *gin.Context) {
	var input signInEmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusUnauthorized, errIncorrectInputData(err.Error()))
		return
	}

	var orgID = c.MustGet("claims_org_id").(uint)

	//return employee model if find
	empl, err := s.repo.Employees.SignIn(input.ID, input.Password, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusUnauthorized, errRecordNotFound())
			return
		}
		NewResponse(c, http.StatusUnauthorized, errUnknownServer(err.Error()))
		return
	}

	//create new claims
	claims := authjwt.EmployeeClaims{
		OrganizationID: orgID,
		OutletID:       empl.OutletID,
		EmployeeID:     empl.ID,
		Role:           empl.Role,
	}

	token, err := s.authjwt.SignInEmployee(&claims)
	if err != nil {
		NewResponse(c, http.StatusUnauthorized, errUnknownServer(err.Error()))
		return
	}

	output := signInEmployeeOutput{Token: token}
	NewResponse(c, http.StatusOK, output)
}

type sendCodeInputQuery struct {
	Type  string `form:"type" binding:"required"`
	Email string `form:"email" binding:"required"`
}

//@Summary Отправка кода подтверждения почты
//@param type query string false "`org` or `employee`"
//@param email query string false "адрес на который будет отправлено письмо (например: email@exmp.ru)"
//@Success 200 {object} object "возвращает пустой объект"
//@Router /auth/sendCode [get]
func (s *AuthorizationService) SendCode(c *gin.Context) {
	var inputQ sendCodeInputQuery
	if err := c.ShouldBindQuery(&inputQ); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	switch inputQ.Type {
	case "org":
		{
			//проверка на сущ. email в БД
			if ok, err := s.repo.Organizations.EmailExists(inputQ.Email); err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
				return
			} else if !ok {
				NewResponse(c, http.StatusBadRequest, errEmailNotFound())
				return
			}

			//отправка письма с ссылкой для подтверждения
			if err := s.mailagent.SendTemplate(inputQ.Email, "confirm_code.html", mailagent.Value{
				"code":     s.strcode.Encode(inputQ.Email),
				"host":     config.Env.Host,
				"port":     config.Env.Port,
				"protocol": config.Env.Protocol,
				"type":     "org",
			}); err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknownServer("error on send email"))
				return
			}
			NewResponse(c, http.StatusOK, nil)
		}
	default:
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData())
	}

}

type confirmCodeInputQuery struct {
	Type string `form:"type" binding:"required"`
	Code string `form:"code" binding:"required"`
}

//@Summary Проверка кода подтверждения
//@param type query string false "`org` or `employee`"
//@param code query string false "адрес на который будет отправлено письмо (например: email@exmp.ru)"
//@Success 200 {object} object "возвращает пустой объект"
//@Router /auth/confirmCode [get]
func (s *AuthorizationService) ConfirmCode(c *gin.Context) {
	var inputQ confirmCodeInputQuery
	if err := c.ShouldBindQuery(&inputQ); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	switch inputQ.Type {
	case "org":
		{
			email, err := s.strcode.Decode(inputQ.Code)
			if err != nil {
				NewResponse(c, http.StatusBadRequest, errIncorrectConfirmCode(err.Error()))
				return
			}

			if err := s.repo.Organizations.ConfirmEmailTrue(email); err != nil {
				NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
				return
			}
		}
	default:
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData())
	}
}
