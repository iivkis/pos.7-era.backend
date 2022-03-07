package myservice

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type EmployeesService struct {
	repo *repository.Repository
}

type EmployeeOutputModel struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	RoleID   int    `json:"role_id"`
	Online   bool   `json:"online"`
	OutletID uint   `json:"outlet_id"`
}

func newEmployeesService(repo *repository.Repository) *EmployeesService {
	return &EmployeesService{
		repo: repo,
	}
}

type EmployeesGetAllForOrgInputQuery struct {
	OutletID uint `form:"outlet_id"`
}
type EmployeesGetAllForOrgOutput []EmployeeOutputModel

//@Summary Список всех сотрудников организации
//@Description Метод позволяет получить список всех сотрудников организации
//@Produce json
//@Success 200 {object} EmployeesGetAllForOrgOutput "Возвращает массив сотрудников"
//@Failure 500 {object} serviceError
//@Router /employees [get]
func (s *EmployeesService) GetAllForOrg(c *gin.Context) {
	var query EmployeesGetAllForOrgInputQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	employees, err := s.repo.Employees.FindAllByOrgID(c.MustGet("claims_org_id"), query.OutletID)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := make(EmployeesGetAllForOrgOutput, len(employees))
	for i, employee := range employees {
		if err != nil {
			NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
			return
		}
		output[i] = EmployeeOutputModel{
			ID:       employee.ID,
			Name:     employee.Name,
			Role:     employee.Role,
			RoleID:   employee.GetRoleID(),
			Online:   employee.Online,
			OutletID: employee.OutletID,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

type EmployeeUpdateFieldsInput struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

//@Summary Позволяет обновить поля сотрудника
//@param type body EmployeeUpdateFieldsInput false "Принимаемый объект"
//@Accept json
//@Produce json
//@Success 200 {object} object "возвращает пустой объект"
//@Failure 400 {object} serviceError
//@Router /employees/:id [put]
func (s *EmployeesService) UpdateFields(c *gin.Context) {
	var input EmployeeUpdateFieldsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	m := &repository.EmployeeModel{
		Name:     input.Name,
		Password: input.Password,
	}

	if err := s.repo.Employees.Updates(m, c.Param("id"), c.MustGet("claims_outlet_id"), c.MustGet("claims_role").(string)); err != nil {
		if errors.Is(err, repository.ErrOnlyNumCanBeInPassword) || errors.Is(err, repository.ErrUndefinedRole) {
			NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound(err.Error()))
		} else if errors.Is(err, repository.ErrPermissionDenided) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
		} else {
			NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		}
		return
	}
}

//@Summary Позволяет удалить сотрудника
//@Accept json
//@Produce json
//@Success 200 {object} object "возвращает пустой объект"
//@Failure 400 {object} serviceError
//@Router /employees/:id [delete]
func (s *EmployeesService) Delete(c *gin.Context) {
	if err := s.repo.Employees.Delete(c.Param("id"), c.MustGet("claims_outlet_id"), c.MustGet("claims_role").(string)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound(err.Error()))
		} else if errors.Is(err, repository.ErrPermissionDenided) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
		} else {
			NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		}
		return
	}
}
