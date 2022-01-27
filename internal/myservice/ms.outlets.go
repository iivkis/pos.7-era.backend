package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

type OutletsService interface {
	Create(c *gin.Context)
	GetAll(c *gin.Context)
}

type outlets struct {
	repo repository.Repository
}

type outletOutputModel struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func newOutletsService(repo repository.Repository) *outlets {
	return &outlets{
		repo: repo,
	}
}

type createOutletInput struct {
	Name string `json:"name" binding:"min=3,max=100"`
}

//@Summary Добавить торговую точку
//@Description Метод позволяет добавить торговую точку
//@Param json body createOutletInput true "Объект для добавления торговой точки."
//@Accept json
//@Produce json
//@Success 200 {object} object "Возвращает пустой объект"
//@Failure 500 {object} serviceError
//@Router /outlets [post]
func (s *outlets) Create(c *gin.Context) {
	var input createOutletInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	model := repository.OutletModel{
		Name:  input.Name,
		OrgID: c.MustGet("claims_org_id").(uint),
	}

	if err := s.repo.Outlets.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, nil)
}

type getAllOutletsOutput []outletOutputModel

//@Summary Список всех торговых точек
//@Description Метод позволяет получить список всех торговых точек
//@Produce json
//@Success 200 {object} getAllOutletsOutput "Возвращает массив торговых точек"
//@Failure 500 {object} serviceError
//@Router /outlets [get]
func (s *outlets) GetAll(c *gin.Context) {
	orgID := c.MustGet("claims_org_id").(uint)

	outlets, err := s.repo.Outlets.GetAll(orgID)
	if err != nil {
		if dberr, ok := isDatabaseError(err); ok {
			NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(dberr.Error()))
			return
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownServer(err.Error()))
		return
	}

	output := make(getAllOutletsOutput, len(outlets))

	for i, outlet := range outlets {
		output[i] = outletOutputModel{
			ID:   outlet.ID,
			Name: outlet.Name,
		}
	}

	NewResponse(c, http.StatusOK, output)
}
