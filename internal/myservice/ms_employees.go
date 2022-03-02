package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type EmployeesService struct {
	repo *repository.Repository
}

type employeeOutputModel struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Online   bool   `json:"online"`
	OutletID uint   `json:"outlet_id"`
}

func newEmployeesService(repo *repository.Repository) *EmployeesService {
	return &EmployeesService{
		repo: repo,
	}
}

type getAllEmployeesOutput []employeeOutputModel

//@Summary Список всех сотрудников организации
//@Description Метод позволяет получить список всех сотрудников организации
//@Produce json
//@Success 200 {object} getAllEmployeesOutput "Возвращает массив сотрудников"
//@Failure 500 {object} serviceError
//@Router /employees [get]
func (s *EmployeesService) GetAllForOrg(c *gin.Context) {
	employees, err := s.repo.Employees.FindAllByOrgID(c.MustGet("claims_org_id"))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := make(getAllEmployeesOutput, len(employees))
	for i, employee := range employees {
		if err != nil {
			NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
			return
		}
		output[i] = employeeOutputModel{
			ID:       employee.ID,
			Name:     employee.Name,
			Role:     employee.Role,
			Online:   employee.Online,
			OutletID: employee.OutletID,
		}
	}

	NewResponse(c, http.StatusOK, output)
}
