package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type inventoryListResponeModel struct {
	ID                 uint `json:"id" mapstructure:"id"`
	InventoryHistoryID uint `json:"inventory_history_id" mapstructure:"inventory_history_id"`
	IngredientID       uint `json:"ingredient_id" mapstructure:"ingredient_id"`
	OutletID           uint `json:"outletID" mapstructure:"outletID"`

	OldCount  float64 `json:"old_count" mapstructure:"old_count"`
	NewCount  float64 `json:"new_count" mapstructure:"new_count"`
	LossPrice float64 `json:"price" mapstructure:"price"`
}

type inventoryList struct {
	repo *repository.Repository
}

func newInventoryList(repo *repository.Repository) *inventoryList {
	return &inventoryList{
		repo: repo,
	}
}

type inventoryListCreateBody struct {
	InventoryHistoryID uint    `json:"inventory_history_id"`
	IngredientID       uint    `json:"ingredient_id"`
	NewCount           float64 `json:"new_count"`
}

// @Summary Добавить отчет об ингредиенте
// @Param type body inventoryListCreateBody false "Принимаемый объект"
// @Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
// @Router /inventoryList [post]
func (s *inventoryList) Create(c *gin.Context) {
	var input inventoryListCreateBody
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	if !s.repo.Ingredients.Exists(&repository.IngredientModel{ID: input.IngredientID, OutletID: claims.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `ingredient` with this `id in outlet`"))
		return
	}

	if !s.repo.InventoryHistory.Exists(&repository.InventoryHistoryModel{Model: gorm.Model{ID: input.InventoryHistoryID}, OutletID: claims.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `inventoryHistory` with this `id` in outlet"))
		return
	}

	ingredient, err := s.repo.Ingredients.FindFirts(&repository.IngredientModel{ID: input.IngredientID})
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}

	model := &repository.InventoryListModel{
		OldCount:           ingredient.Count,
		NewCount:           input.NewCount,
		IngredientID:       input.IngredientID,
		InventoryHistoryID: input.InventoryHistoryID,
		OutletID:           claims.OutletID,
		OrgID:              claims.OrganizationID,
		LossPrice:          0,
	}

	if model.OldCount != model.NewCount {
		model.LossPrice = (model.OldCount - model.NewCount) * ingredient.PurchasePrice

		if err := s.repo.Ingredients.Updates(
			&repository.IngredientModel{ID: ingredient.ID},
			&repository.IngredientModel{Count: model.NewCount},
		); err != nil {
			NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
			return
		}
	}

	if err := s.repo.InventoryList.Create(model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type inventoryListGetAllQuery struct {
	InventoryHistoryID uint `form:"inventory_history_id"`
}
type inventoryListGetAllResponse []inventoryListResponeModel

// @Summary Получить всю историю инвернтаризации
// @param type query inventoryListGetAllQuery false "query"
// @Success 200 {object} inventoryListGetAllResponse "возвращаемый объект"
// @Router /inventoryList [get]
func (s *inventoryList) GetAll(c *gin.Context) {
	var query inventoryListGetAllQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.InventoryListModel{
		OrgID:              claims.OrganizationID,
		OutletID:           claims.OutletID,
		InventoryHistoryID: query.InventoryHistoryID,
	}

	if claims.HasRole(repository.R_OWNER) {
		if stdQuery.OrgID != 0 && s.repo.Invitation.Exists(&repository.InvitationModel{OrgID: claims.OrganizationID, AffiliateOrgID: stdQuery.OrgID}) {
			where.OrgID = stdQuery.OrgID
		}
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	invetoryList, err := s.repo.InventoryList.Find(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	var output inventoryListGetAllResponse = make(inventoryListGetAllResponse, len(*invetoryList))
	for i, item := range *invetoryList {
		output[i] = inventoryListResponeModel{
			ID:                 item.ID,
			OldCount:           item.OldCount,
			NewCount:           item.NewCount,
			LossPrice:          item.LossPrice,
			IngredientID:       item.IngredientID,
			InventoryHistoryID: item.InventoryHistoryID,
			OutletID:           item.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}
