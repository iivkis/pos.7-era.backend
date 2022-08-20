package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type employees struct {
	repo *repository.Repository
}

type EmployeeOutputModel struct {
	ID       uint   `json:"id" mapstructure:"id"`
	Name     string `json:"name" mapstructure:"name"`
	Role     string `json:"role" mapstructure:"role"`
	RoleID   int    `json:"role_id" mapstructure:"role_id"`
	Online   bool   `json:"online" mapstructure:"online"`
	OutletID uint   `json:"outlet_id" mapstructure:"outlet_id"`
}

func newEmployees(repo *repository.Repository) *employees {
	return &employees{
		repo: repo,
	}
}

type employeesGetAllResponse []EmployeeOutputModel

// @Summary Список всех сотрудников организации
// @Description Метод позволяет получить список всех сотрудников организации
// @Produce json
// @Success 200 {object} employeesGetAllResponse "Возвращает массив сотрудников"
// @Failure 500 {object} serviceError
// @Router /employees [get]
func (s *employees) GetAll(c *gin.Context) {
	claims, stdQuery := mustGetOrganizationClaims(c), mustGetStandartQuery(c)

	where := &repository.EmployeeModel{
		OrgID: claims.OrganizationID,
	}

	if stdQuery.OutletID != 0 {
		if ok, _ := s.repo.HasAccessToOutlet(claims.OrganizationID, stdQuery.OutletID); ok {
			where.OrgID = 0
			where.OutletID = stdQuery.OutletID
		} else {
			NewResponse(c, http.StatusForbidden, errPermissionDenided())
			return
		}
	}

	employees, err := s.repo.Employees.Find(where)

	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := make(employeesGetAllResponse, len(*employees))
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

type employeeUpdateBody struct {
	Name     string `json:"name"`
	Password string `json:"password" binding:"max=6"`
	RoleID   int    `json:"role_id" binding:"min=0"`
}

// @Summary Обновить поля сотрудника
// @Param type body employeeUpdateBody false "object"
// @Success 200 {object} object "object"
// @Router /employees/:id [put]
func (s *employees) Update(c *gin.Context) {
	employeeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	var body employeeUpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	if !repository.RoleIsExists(repository.RoleIDToName(body.RoleID)) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined role"))
		return
	}

	claims := mustGetEmployeeClaims(c)

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
				Name:     body.Name,
				Password: body.Password,
			}
		} else if editedEmployee.HasRole(repository.R_DIRECTOR, repository.R_ADMIN, repository.R_CASHIER) {
			updatedFields = &repository.EmployeeModel{
				Name:     body.Name,
				Password: body.Password,
				Role:     repository.RoleIDToName(body.RoleID),
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
				Password: body.Password,
			}
		} else if editedEmployee.HasRole(repository.R_ADMIN, repository.R_CASHIER) {
			updatedFields = &repository.EmployeeModel{
				Name:     body.Name,
				Password: body.Password,
				Role:     repository.RoleIDToName(body.RoleID),
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
				Password: body.Password,
			}
		} else if editedEmployee.HasRole(repository.R_CASHIER) {
			updatedFields = &repository.EmployeeModel{
				Name:     body.Name,
				Password: body.Password,
				Role:     repository.RoleIDToName(body.RoleID),
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

// @Summary Позволяет удалить сотрудника
// @Accept json
// @Produce json
// @Success 200 {object} object "возвращает пустой объект"
// @Failure 400 {object} serviceError
// @Router /employees/:id [delete]
func (s *employees) Delete(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

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
