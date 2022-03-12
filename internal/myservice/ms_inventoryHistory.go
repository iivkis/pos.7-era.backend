package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type InventoryHistoryOutputModel struct {
	ID         uint  `json:"id"`
	Date       int64 `json:"date"`        //unixmilli
	EmployeeID uint  `json:"employee_id"` //сотрудник, который делал инветаризацию
	OutletID   uint  `json:"outlet_id"`
}

type InventoryHistoryService struct {
	repo *repository.Repository
}

func newInventoryHistoryService(repo *repository.Repository) *InventoryHistoryService {
	return &InventoryHistoryService{
		repo: repo,
	}
}

type InventoryHistoryCreateInput struct {
	Date int64 //unixmilli
}

//@Summary Добавить новую историю инвентаризации
//@param type body InventoryHistoryCreateInput false "Принимаемый объект"
//@Accept json
//@Produce json
//@Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
//@Failure 400 {object} serviceError
//@Router /invetoryHistory [post]
func (s *InventoryHistoryService) Create(c *gin.Context) {
	var input InventoryHistoryCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	model := &repository.InventoryHistoryModel{
		Date:       input.Date,
		EmployeeID: claims.EmployeeID,
		OutletID:   claims.OutletID,
		OrgID:      claims.OrganizationID,
	}

	if err := s.repo.InventoryHistory.Create(model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type InventoryHistoryGetAllOutput []InventoryHistoryOutputModel

//@Summary Получить всю историю инвернтаризации
//@Accept json
//@Produce json
//@Success 200 {object} InventoryHistoryGetAllOutput "возвращаемый объект"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /invetoryHistory [get]
func (s *InventoryHistoryService) GetAll(c *gin.Context) {
	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.InventoryHistoryModel{
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	invetoryHistoryList, err := s.repo.InventoryHistory.Find(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output InventoryHistoryGetAllOutput = make(InventoryHistoryGetAllOutput, len(*invetoryHistoryList))
	for i, item := range *invetoryHistoryList {
		output[i] = InventoryHistoryOutputModel{
			ID:         item.ID,
			Date:       item.Date,
			EmployeeID: item.EmployeeID,
			OutletID:   item.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}
