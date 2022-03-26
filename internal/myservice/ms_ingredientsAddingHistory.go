package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type IngredientsAddingHistoryOutputModel struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Count         float64 `json:"count"`
	MeasureUnit   int     `json:"measure_unit"`
	PurchasePrice float64 `json:"purchase_price"`
	OutletID      uint    `json:"outlet_id"`
}

type IngredientsAddingHistoryService struct {
	repo *repository.Repository
}

func newIngredientsAddingHistoryService(repo *repository.Repository) *IngredientsAddingHistoryService {
	return &IngredientsAddingHistoryService{
		repo: repo,
	}
}

type IngredientsAddingHistoryCreateInput struct {
	Name          string  `json:"name" binding:"required"`
	Count         float64 `json:"count" binding:"min=0"`
	PurchasePrice float64 `json:"purchase_price" binding:"min=0"`
	MeasureUnit   int     `json:"measure_unit" binding:"min=1,max=3"`
}

//@Summary Добавить новый ингредиент в точку
//@param type body IngredientCreateInput false "Принимаемый объект"
//@Accept json
//@Produce json
//@Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /ingredients [post]
func (s *IngredientsAddingHistoryService) Create(c *gin.Context) {
	var input IngredientCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	model := repository.IngredientModel{
		Name:          input.Name,
		Count:         input.Count,
		PurchasePrice: input.PurchasePrice,
		MeasureUnit:   input.MeasureUnit,
		OutletID:      claims.OutletID,
		OrgID:         claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			model.OutletID = stdQuery.OutletID
		}
	}

	if err := s.repo.Ingredients.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type IngredientsAddingHistorytGetAllOutput []IngredientOutputModel

//@Summary Получить все ингредиенты точки
//@Accept json
//@Produce json
//@Success 200 {object} IngredientGetAllOutput "возвращает все ингредиенты текущей точки"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /ingredients [get]
func (s *IngredientsAddingHistoryService) GetAll(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)
	stdQuery := mustGetStdQuery(c)

	where := &repository.IngredientModel{
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	ingredients, err := s.repo.Ingredients.Find(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output IngredientGetAllOutput = make(IngredientGetAllOutput, len(*ingredients))
	for i, ingredient := range *ingredients {
		output[i] = IngredientOutputModel{
			ID:            ingredient.ID,
			Name:          ingredient.Name,
			Count:         ingredient.Count,
			MeasureUnit:   ingredient.MeasureUnit,
			PurchasePrice: ingredient.PurchasePrice,
			OutletID:      ingredient.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}
