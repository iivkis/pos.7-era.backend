package myservice

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
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

type EmployeesGetAllOutput []EmployeeOutputModel

//@Summary Список всех сотрудников организации
//@Description Метод позволяет получить список всех сотрудников организации
//@Produce json
//@Success 200 {object} EmployeesGetAllOutput "Возвращает массив сотрудников"
//@Failure 500 {object} serviceError
//@Router /employees [get]
func (s *EmployeesService) GetAll(c *gin.Context) {
	claims, stdQuery := mustGetOrganizationClaims(c), mustGetStdQuery(c)

	employees, err := s.repo.Employees.Find(&repository.EmployeeModel{
		OrgID:    claims.OrganizationID,
		OutletID: stdQuery.OutletID,
	})

	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := make(EmployeesGetAllOutput, len(*employees))
	for i, employee := range *employees {
		if err != nil {
			NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
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
	Password string `json:"password" binding:"max=6"`
	RoleID   int    `json:"role_id"`
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

	if input.RoleID != 0 && !repository.RoleIsExists(repository.RoleIDToName(input.RoleID)) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined role"))
		return
	}

	employeeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := c.MustGet("claims").(*authjwt.EmployeeClaims)

	editedEmployee, err := s.repo.Employees.FindFirst(&repository.EmployeeModel{Model: gorm.Model{ID: uint(employeeID)}, OrgID: claims.OrganizationID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound())
		} else {
			NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		}
		return
	}

	var updatedFields *repository.EmployeeModel

	switch claims.Role {
	case repository.R_OWNER:
		if claims.EmployeeID == editedEmployee.ID {
			updatedFields = &repository.EmployeeModel{
				Name:     input.Name,
				Password: input.Password,
			}
		} else if editedEmployee.HasRole(repository.R_DIRECTOR, repository.R_ADMIN, repository.R_CASHIER) {
			updatedFields = &repository.EmployeeModel{
				Name:     input.Name,
				Password: input.Password,
				Role:     repository.RoleIDToName(input.RoleID),
			}
		} else {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}

		if updatedFields.Role != "" && !updatedFields.HasRole(repository.R_DIRECTOR, repository.R_ADMIN, repository.R_CASHIER) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}

	case repository.R_DIRECTOR:
		if claims.EmployeeID == editedEmployee.ID {
			updatedFields = &repository.EmployeeModel{
				Password: input.Password,
			}
		} else if editedEmployee.HasRole(repository.R_ADMIN, repository.R_CASHIER) {
			updatedFields = &repository.EmployeeModel{
				Name:     input.Name,
				Password: input.Password,
				Role:     repository.RoleIDToName(input.RoleID),
			}
		} else {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}

		if updatedFields.Role != "" && !updatedFields.HasRole(repository.R_ADMIN, repository.R_CASHIER) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}

	case repository.R_ADMIN:
		//только в своей точке
		if claims.OutletID != editedEmployee.OutletID {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}

		if claims.EmployeeID == editedEmployee.ID {
			updatedFields = &repository.EmployeeModel{
				Password: input.Password,
			}
		} else if editedEmployee.HasRole(repository.R_CASHIER) {
			updatedFields = &repository.EmployeeModel{
				Name:     input.Name,
				Password: input.Password,
				Role:     repository.RoleIDToName(input.RoleID),
			}
		} else {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}

		if updatedFields.Role != "" && !updatedFields.HasRole(repository.R_ADMIN, repository.R_CASHIER) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}
	default:
		NewResponse(c, http.StatusBadRequest, errPermissionDenided())
		return
	}

	if err := s.repo.Employees.Updates(updatedFields, &repository.EmployeeModel{Model: gorm.Model{ID: uint(employeeID)}}); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

//@Summary Позволяет удалить сотрудника
//@Accept json
//@Produce json
//@Success 200 {object} object "возвращает пустой объект"
//@Failure 400 {object} serviceError
//@Router /employees/:id [delete]
func (s *EmployeesService) Delete(c *gin.Context) {
	claims := c.MustGet("claims").(*authjwt.EmployeeClaims)

	employeeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	deletedEmployee, err := s.repo.Employees.FindFirst(&repository.EmployeeModel{Model: gorm.Model{ID: uint(employeeID)}, OrgID: claims.OrganizationID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound())
		} else {
			NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		}
		return
	}

	switch claims.Role {
	case repository.R_OWNER:
		if !deletedEmployee.HasRole(repository.R_DIRECTOR, repository.R_ADMIN, repository.R_CASHIER) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}
	case repository.R_DIRECTOR:
		if !deletedEmployee.HasRole(repository.R_ADMIN, repository.R_CASHIER) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}
	case repository.R_ADMIN:
		if claims.OutletID != deletedEmployee.OutletID || !deletedEmployee.HasRole(repository.R_CASHIER) {
			NewResponse(c, http.StatusBadRequest, errPermissionDenided())
			return
		}
	}

	if err := s.repo.Employees.Delete(&repository.EmployeeModel{Model: gorm.Model{ID: uint(employeeID)}, OrgID: claims.OrganizationID}); err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}
