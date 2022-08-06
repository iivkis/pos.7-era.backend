package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type InventoryHistoryResponseModel struct {
	ID         uint `json:"id" mapstructure:"id"`
	EmployeeID uint `json:"employee_id" mapstructure:"employee_id"`
	OutletID   uint `json:"outlet_id" mapstructure:"outlet_id"`

	Date int64 `json:"date" mapstructure:"date"` //unixmilli
}

type inventoryHistory struct {
	repo *repository.Repository
}

func newInventoryHistory(repo *repository.Repository) *inventoryHistory {
	return &inventoryHistory{
		repo: repo,
	}
}

type inventoryHistoryCreateBody struct {
	Date int64 `json:"date"` //unixmilli, дата инвенторизации
}

// @Summary Добавить историю инвентаризации
// @param type body InventoryHistoryCreateInput false "object"
// @Success 201 {object} DefaultOutputModel "id"
// @Router /inventoryHistory [post]
func (s *inventoryHistory) Create(c *gin.Context) {
	var body inventoryHistoryCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	model := &repository.InventoryHistoryModel{
		EmployeeID: claims.EmployeeID,
		OutletID:   claims.OutletID,
		OrgID:      claims.OrganizationID,
		Date:       body.Date,
	}

	if err := s.repo.InventoryHistory.Create(model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type inventoryHistoryGetAllQuery struct {
	Start uint64 `form:"start"` //in unixmilli
	End   uint64 `form:"end"`   //in unixmilli
}

type inventoryHistoryGetAllResponse []InventoryHistoryResponseModel

// @Summary Получить всю историю инвернтаризации
// @param type query inventoryHistoryGetAllQuery false "Принимаемый объект"
// @Success 200 {object} inventoryHistoryGetAllResponse "истрия инвенторизации "
// @Router /inventoryHistory [get]
func (s *inventoryHistory) GetAll(c *gin.Context) {
	var query inventoryHistoryGetAllQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.InventoryHistoryModel{
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER) {
		if stdQuery.OrgID != 0 && s.repo.Invitation.Exists(&repository.InvitationModel{OrgID: claims.OrganizationID, AffiliateOrgID: stdQuery.OrgID}) {
			where.OrgID = stdQuery.OrgID
		}
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	invetoryHistoryList, err := s.repo.InventoryHistory.FindWithPeriod(where, query.Start, query.End)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	var output inventoryHistoryGetAllResponse = make(inventoryHistoryGetAllResponse, len(*invetoryHistoryList))
	for i, item := range *invetoryHistoryList {
		output[i] = InventoryHistoryResponseModel{
			ID:         item.ID,
			Date:       item.Date,
			EmployeeID: item.EmployeeID,
			OutletID:   item.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}
