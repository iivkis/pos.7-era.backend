package myservice

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type InventoryListOutputModel struct {
	ID                 uint    `json:"id"`
	OldCount           float64 `json:"old_count"`
	NewCount           float64 `json:"new_count"`
	LossPrice          float64 `json:"price"`
	IngredientID       uint    `json:"ingredient_id"`
	InventoryHistoryID uint    `json:"inventory_history_id"`
	OutletID           uint    `json:"outletID"`
}

type InventoryListService struct {
	repo *repository.Repository
}

func newInventoryListService(repo *repository.Repository) *InventoryListService {
	return &InventoryListService{
		repo: repo,
	}
}

type InventoryListCreateInput struct {
	NewCount           float64 `json:"new_count"`
	IngredientID       uint    `json:"ingredient_id"`
	InventoryHistoryID uint    `json:"inventory_history_id"`
}

//@Summary Добавить InventoryList
//@param type body InventoryListCreateInput false "Принимаемый объект"
//@Accept json
//@Produce json
//@Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
//@Failure 400 {object} serviceError
//@Router /inventoryList [post]
func (s *InventoryListService) Create(c *gin.Context) {
	var input InventoryListCreateInput
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
		NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
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
			NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
			return
		}
	}

	if err := s.repo.InventoryList.Create(model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type InventoryListGetAllQuery struct {
	InventoryHistoryID uint `form:"inventory_history_id"`
}
type InventoryListGetAllOutput []InventoryListOutputModel

//@Summary Получить всю историю инвернтаризации
//@param type query InventoryListGetAllQuery false "Принимаемый объект"
//@Accept json
//@Produce json
//@Success 200 {object} InventoryListGetAllOutput "возвращаемый объект"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /inventoryList [get]
func (s *InventoryListService) GetAll(c *gin.Context) {
	var query InventoryListGetAllQuery
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

	fmt.Println(query)

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	invetoryList, err := s.repo.InventoryList.Find(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output InventoryListGetAllOutput = make(InventoryListGetAllOutput, len(*invetoryList))
	for i, item := range *invetoryList {
		output[i] = InventoryListOutputModel{
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
