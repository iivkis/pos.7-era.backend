package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

type IngredientOutputModel struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Count       float64 `json:"count"`
	MeasureUnit int     `json:"measure_unit"`
	OutletID    uint    `json:"outlet_id"`
}

type IngredientsService struct {
	repo *repository.Repository
}

func newIngredientsService(repo *repository.Repository) *IngredientsService {
	return &IngredientsService{
		repo: repo,
	}
}

type IngredientCreateInput struct {
	Name          string  `json:"name" binding:"required"`
	Count         float64 `json:"count"`
	PurchasePrice float64 `json:"purchase_price"`
	MeasureUnit   int     `json:"measure_unit" binding:"min=1,max=3"`
}

func (s *IngredientsService) Create(c *gin.Context) {
	var input IngredientCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	ingredient := repository.IngredientModel{
		Name:          input.Name,
		Count:         input.Count,
		PurchasePrice: input.PurchasePrice,
		MeasureUnit:   input.MeasureUnit,
		OutletID:      c.MustGet("claims_outlet_id").(uint),
		OrgID:         c.MustGet("claims_org_id").(uint),
	}

	if err := s.repo.Ingredients.Create(&ingredient); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

type IngredientGetAllOutput []IngredientOutputModel

func (s *IngredientsService) GetAllForOrg(c *gin.Context) {
	ingredients, err := s.repo.Ingredients.GetAllByOrgID(c.MustGet("claims_org_id").(uint))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output IngredientGetAllOutput = make(IngredientGetAllOutput, len(ingredients))
	for i, ingredient := range ingredients {
		output[i] = IngredientOutputModel{
			ID:          ingredient.ID,
			Name:        ingredient.Name,
			Count:       ingredient.Count,
			MeasureUnit: ingredient.MeasureUnit,
			OutletID:    ingredient.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}

type IngredientUpdateInput struct {
	Name          string  `json:"name"`
	Count         float64 `json:"count"`
	PurchasePrice float64 `json:"purchase_price"`
	MeasureUnit   int     `json:"measure_unit"`
}

func (s *IngredientsService) UpdateModel(c *gin.Context) {
	var input IngredientUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	ingredient := repository.IngredientModel{
		Name:          input.Name,
		PurchasePrice: input.PurchasePrice,
		Count:         input.Count,
		MeasureUnit:   input.MeasureUnit,
	}

	if err := s.repo.Ingredients.Update(&ingredient, c.Param("id"), c.MustGet("claims_outlet_id").(uint)); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

func (s *IngredientsService) Delete(c *gin.Context) {
	if err := s.repo.Ingredients.Delete(c.Param("id"), c.MustGet("claims_outlet_id")); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}
