package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type ingredientsAddingHistoryResponseModel struct {
	ID           uint `json:"id" mapstructure:"id"`
	IngredientID uint `json:"ingredient_id" mapstructure:"ingredient_id"`
	EmployeeID   uint `json:"employee_id" mapstructure:"employee_id"`
	OutletID     uint `json:"outlet_id" mapstructure:"outlet_id"`

	Count  float64 `json:"count" mapstructure:"count"`   //кол-во продукта, который не сходится
	Total  float64 `json:"total" mapstructure:"total"`   //сумма, на которую не сходится
	Status int     `json:"status" mapstructure:"status"` // 1 - инвенторизация

	Date int64 `json:"date" mapstructure:"date"` //unixmilli
}

type ingredientsAddingHistory struct {
	repo *repository.Repository
}

func newIngredientsAddingHistory(repo *repository.Repository) *ingredientsAddingHistory {
	return &ingredientsAddingHistory{
		repo: repo,
	}
}

type ingredientsAddingHistoryCreateBody struct {
	IngredientID uint `json:"ingredient_id" binding:"min=1"`

	Date   int64   `json:"date"`                         //unixmilli
	Count  float64 `json:"count"`                        //кол-во продукта, который не сходится
	Total  float64 `json:"total"`                        //сумма, на которую не сходится
	Status int     `json:"status" binding:"min=1,max=2"` // 1 - инвенторизация
}

// @Summary Добавить новый отчёт об ингредиентах
// @Param type body ingredientsAddingHistoryCreateBody false "object"
// @Success 201 {object} DefaultOutputModel "id"
// @Router /ingredients.History [post]
func (s *ingredientsAddingHistory) Create(c *gin.Context) {
	var input ingredientsAddingHistoryCreateBody
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStandartQuery(c)

	model := repository.IngredientsAddingHistoryModel{
		Count:        input.Count,
		Total:        input.Total,
		Status:       input.Status,
		IngredientID: input.IngredientID,
		Date:         input.Date,
		EmployeeID:   claims.EmployeeID,
		OutletID:     claims.OutletID,
		OrgID:        claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			model.OutletID = stdQuery.OutletID
		}
	}

	if !s.repo.Ingredients.Exists(&repository.IngredientModel{ID: model.IngredientID, OutletID: model.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `ingredient_id` in outlet with this id"))
		return
	}

	if err := s.repo.IngredientsAddingHistory.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type ingredientsAddingHistorytGetAllQuery struct {
	Start uint64 `form:"start"` //in unixmilli
	End   uint64 `form:"end"`   //in unixmilli
}
type ingredientsAddingHistorytGetAllResponse []ingredientsAddingHistoryResponseModel

// @Summary Получить историю добавления ингредиентов
// @Success 200 {object} ingredientsAddingHistorytGetAllResponse "история"
// @Router /ingredients.History [get]
func (s *ingredientsAddingHistory) GetAll(c *gin.Context) {
	var query ingredientsAddingHistorytGetAllQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}
	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStandartQuery(c)

	where := &repository.IngredientsAddingHistoryModel{
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

	histories, err := s.repo.IngredientsAddingHistory.FindWithPeriod(where, query.Start, query.End)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	var output = make(ingredientsAddingHistorytGetAllResponse, len(*histories))
	for i, item := range *histories {
		output[i] = ingredientsAddingHistoryResponseModel{
			ID:           item.ID,
			Count:        item.Count,
			Total:        item.Total,
			Status:       item.Status,
			Date:         item.Date,
			IngredientID: item.IngredientID,
			EmployeeID:   item.EmployeeID,
			OutletID:     item.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}
