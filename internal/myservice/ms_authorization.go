package myservice

import (
	"errors"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"github.com/iivkis/pos.7-era.backend/pkg/mailagent"
	"github.com/iivkis/strcode"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthorizationService struct {
	repo      *repository.Repository
	strcode   *strcode.Strcode
	mailagent *mailagent.MailAgent
	authjwt   *authjwt.AuthJWT
}

func newAuthorizationService(repo *repository.Repository, strcode *strcode.Strcode, mailagent *mailagent.MailAgent, authjwt *authjwt.AuthJWT) *AuthorizationService {
	return &AuthorizationService{
		repo:      repo,
		strcode:   strcode,
		mailagent: mailagent,
		authjwt:   authjwt,
	}
}

type SignUpOrgInput struct {
	Name       string `json:"name" binding:"required,min=3,max=50"`
	Email      string `json:"email" binding:"required,min=3,max=50"`
	Password   string `json:"password" binding:"required,min=6,max=45"`
	InviteCode string `json:"invite_code" binding:"max=9"`
}

//@Summary Регистрация организации
//@Description Метод позволяет зарегистрировать организацию
//@Param json body SignUpOrgInput true "Объект для регитсрации огранизации."
//@Accept json
//@Produce json
//@Success 201 {object} object "Возвращаемый объект при регистрации огранизации"
//@Failure 401 {object} serviceError
//@Router /auth/signUp.Org [post]
func (s *AuthorizationService) SignUpOrg(c *gin.Context) {
	var input SignUpOrgInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	//validate email
	if _, err := mail.ParseAddress(input.Email); err != nil {
		NewResponse(c, http.StatusUnauthorized, errIncorrectEmail(err.Error()))
		return
	}

	//chek invite
	if input.InviteCode != "" && !s.repo.Invitation.Exists(&repository.InvitationModel{Code: input.InviteCode}) {
		NewResponse(c, http.StatusBadRequest, errRecordNotFound("unknown invite"))
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
				return
			}
		}
		NewResponse(c, http.StatusUnauthorized, errUnknown(err.Error()))
		return
	}

	//активация инвайта
	if input.InviteCode != "" {
		if err := s.repo.Invitation.Activate(input.InviteCode, orgModel.ID); err != nil {
			NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
			return
		}
	}

	//Создание главной торговой точки
	outletModel := repository.OutletModel{
		Name:  "Главная точка продаж",
		OrgID: orgModel.ID,
	}

	if err := s.repo.Outlets.Create(&outletModel); err != nil {
		NewResponse(c, http.StatusUnauthorized, errUnknown(err.Error()))
		return
	}

	//Создание аккаунта владельца
	employeeOwnerModel := repository.EmployeeModel{
		Name:     "Управление организацией",
		Password: "000000",
		Role:     repository.R_OWNER,
		OrgID:    orgModel.ID,
		OutletID: outletModel.ID,
	}

	if err := s.repo.Employees.Create(&employeeOwnerModel); err != nil {
		NewResponse(c, http.StatusUnauthorized, errUnknown(err.Error()))
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
		NewResponse(c, http.StatusUnauthorized, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, nil)
}

type SignUpEmployeeInput struct {
	Name     string `json:"name" binding:"required,min=2,max=200"`
	Password string `json:"password" binding:"required,min=6,max=6"`
	RoleID   int    `json:"role_id" binding:"min=1"`
}

//@Summary Регистрация сотрудника
//@Description Метод позволяет зарегистрировать ссотрудника. Работает только с токеном организации.
//@Param json body SignUpEmployeeInput true "Объект для регитсрации сотрудника."
//@Accept json
//@Produce json
//@Success 201 {object} object "Возвращаемый объект при регистрации сотрудника"
//@Failure 400 {object} serviceError
//@Router /auth/signUp.Employee [post]
func (s *AuthorizationService) SignUpEmployee(c *gin.Context) {
	var input SignUpEmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)
	stdQuery := mustGetStdQuery(c)

	employeeModel := repository.EmployeeModel{
		Name:     input.Name,
		Password: input.Password,
		Role:     repository.RoleIDToName(input.RoleID),
		OutletID: claims.OutletID,
		OrgID:    claims.OrganizationID,
	}

	if !repository.RoleIsExists(employeeModel.Role) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined role"))
		return
	}

	switch claims.Role {
	case repository.R_OWNER:
		//владелец может создавать только директоров, админов и кассиров
		if !employeeModel.HasRole(repository.R_DIRECTOR, repository.R_ADMIN, repository.R_CASHIER) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}

		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			employeeModel.OutletID = stdQuery.OutletID
		}

	case repository.R_DIRECTOR:
		if !employeeModel.HasRole(repository.R_ADMIN, repository.R_CASHIER) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			employeeModel.OutletID = stdQuery.OutletID
		}

	case repository.R_ADMIN:
		if !employeeModel.HasRole(repository.R_ADMIN, repository.R_CASHIER) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}

	default:
		NewResponse(c, http.StatusBadRequest, errPermissionDenided())
		return
	}

	if err := s.repo.Employees.Create(&employeeModel); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: employeeModel.ID})
}

type SignInOrgInput struct {
	Email    string `json:"email" binding:"required,max=45"`
	Password string `json:"password" binding:"required,max=45"`
}

type SignInOrgOutput struct {
	Token string `json:"token"`
}

//@Summary Вход для организации
//@Description Метод позволяет войти в аккаунт организации.
//@Param json body SignInOrgInput true "Объект для входа в огранизацию."
//@Accept json
//@Produce json
//@Success 200 {object} SignInOrgOutput "Возвращает `jwt токен` при успешной авторизации"
//@Failure 401 {object} serviceError
//@Router /auth/signIn.Org [post]
func (s *AuthorizationService) SignInOrg(c *gin.Context) {
	errUnknown("блят")
	var input SignInOrgInput
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
		NewResponse(c, http.StatusUnauthorized, errUnknown(err.Error()))
		return
	}

	claims := authjwt.OrganizationClaims{
		OrganizationID: org.ID,
	}

	token, err := s.authjwt.SignInOrganization(&claims)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := SignInOrgOutput{Token: token}
	NewResponse(c, http.StatusOK, output)
}

type SignInEmployeeInput struct {
	ID       uint   `json:"id"`
	Password string `json:"password" binding:"required,max=45"`
}

type SignInEmployeeOutput struct {
	Token     string `json:"token"`
	Affiliate bool   `json:"affiliate"` // является ли организация филиалом
}

//@Summary Вход для сотрудника
//@Description Метод позволяет войти в аккаунт сотрудника. Работает только с токеном огранизации.
//@Param json body SignInEmployeeInput true "Объект для входа в огранизацию."
//@Accept json
//@Produce json
//@Success 200 {object} SignInEmployeeOutput "Возвращает `jwt токен` при успешной авторизации"
//@Failure 401 {object} serviceError
//@Router /auth/signIn.Employee [post]
func (s *AuthorizationService) SignInEmployee(c *gin.Context) {
	var input SignInEmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusUnauthorized, errIncorrectInputData(err.Error()))
		return
	}
	claims := mustGetOrganizationClaims(c)

	//return employee model if find
	empl, err := s.repo.Employees.SignIn(input.ID, input.Password, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusUnauthorized, errRecordNotFound())
			return
		}
		NewResponse(c, http.StatusUnauthorized, errUnknown(err.Error()))
		return
	}

	//create new claims
	newEmployeeClaims := authjwt.EmployeeClaims{
		OrganizationID: claims.OrganizationID,
		OutletID:       empl.OutletID,
		EmployeeID:     empl.ID,
		Role:           empl.Role,
	}

	token, err := s.authjwt.SignInEmployee(&newEmployeeClaims)
	if err != nil {
		NewResponse(c, http.StatusUnauthorized, errUnknown(err.Error()))
		return
	}

	output := SignInEmployeeOutput{
		Token:     token,
		Affiliate: s.repo.Invitation.Exists(&repository.InvitationModel{AffiliateOrgID: claims.OrganizationID}),
	}

	NewResponse(c, http.StatusOK, output)
}

type SendCodeInputQuery struct {
	Email string `form:"email" binding:"required"`
}

//@Summary Отправка кода подтверждения почты
//@param email query string false "адрес на который будет отправлено письмо (например: email@exmp.ru)"
//@Success 200 {object} object "возвращает пустой объект"
//@Router /auth/sendCode [get]
func (s *AuthorizationService) SendCode(c *gin.Context) {
	var inputQ SendCodeInputQuery
	if err := c.ShouldBindQuery(&inputQ); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	//проверка на сущ. email в БД
	if !s.repo.Organizations.EmailExists(inputQ.Email) {
		NewResponse(c, http.StatusBadRequest, errEmailNotFound())
		return
	}

	//отправка письма с ссылкой для подтверждения
	if err := s.mailagent.SendTemplate(inputQ.Email, "confirm_code.html", mailagent.Value{
		"code":     s.strcode.Encode(inputQ.Email),
		"host":     config.Env.OutHost,
		"port":     config.Env.OutPort,
		"protocol": config.Env.OutProtocol,
		"type":     "org",
	}); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown("error on send email"))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

type AuthConfirmCodeQuery struct {
	Code string `form:"code" binding:"required"`
}

//@Summary Проверка кода подтверждения
//@param type query AuthConfirmCodeQuery false "адрес на который будет отправлено письмо (например: email@exmp.ru)"
//@Success 200 {object} object "возвращает пустой объект"
//@Router /auth/confirmCode [get]
func (s *AuthorizationService) ConfirmCode(c *gin.Context) {
	var inputQ AuthConfirmCodeQuery
	if err := c.ShouldBindQuery(&inputQ); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	email, err := s.strcode.Decode(inputQ.Code)
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectConfirmCode(err.Error()))
		return
	}

	if err := s.repo.Organizations.SetConfirmEmail(email, true); err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}
}
