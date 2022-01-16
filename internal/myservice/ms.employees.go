package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

type EmployeesService interface {
	GetAll(c *gin.Context)
}

type employees struct {
	repo repository.Repository
}

func newEmployeesService(repo repository.Repository) *employees {
	return &employees{
		repo: repo,
	}
}

type employeeOutputModel struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

//METHODS

type getAllEmployeesOutput []employeeOutputModel

//@Summary Список всех сотрудников
//@Description Метод позволяет получить список всех сотрудников
//@Produce json
//@Success 200 {object} getAllEmployeesOutput "Возвращает массив сотрудников"
//@Failure 500 {object} serviceError
//@Router /employees [get]
func (s *employees) GetAll(c *gin.Context) {
	employees, err := s.repo.Employees.GetAll(c.MustGet("claims_org_id").(uint))
	if err != nil {
		if dberr, ok := isDatabaseError(err); ok {
			NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(dberr.Error()))
			return
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownServer(err.Error()))
		return
	}

	output := make(getAllEmployeesOutput, len(employees))
	for i, employee := range employees {
		output[i] = employeeOutputModel{
			ID:   employee.ID,
			Name: employee.Name,
			Role: employee.Role,
		}
	}

	NewResponse(c, http.StatusOK, output)
}
