package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type IngredientOutputModel struct {
	ID            uint    `json:"id" mapstructure:"id"`
	OutletID      uint    `json:"outlet_id" mapstructure:"outlet_id"`
	Name          string  `json:"name" mapstructure:"name"`
	Count         float64 `json:"count" mapstructure:"count"`
	MeasureUnit   int     `json:"measure_unit" mapstructure:"measure_unit"`
	PurchasePrice float64 `json:"purchase_price" mapstructure:"purchase_price"`
}

type ingredients struct {
	repo *repository.Repository
}

func newIngredients(repo *repository.Repository) *ingredients {
	return &ingredients{
		repo: repo,
	}
}

type ingredientsCreateBody struct {
	Name          string  `json:"name" binding:"required"`
	Count         float64 `json:"count" binding:"min=0"`
	PurchasePrice float64 `json:"purchase_price" binding:"min=0"`
	MeasureUnit   int     `json:"measure_unit" binding:"min=1,max=3"`
}

// @Summary Добавить новый ингредиент в точку
// @param type body ingredientsCreateBody false "Принимаемый объект"
// @Accept json
// @Produce json
// @Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
// @Failure 400 {object} serviceError
// @Failure 500 {object} serviceError
// @Router /ingredients [post]
func (s *ingredients) Create(c *gin.Context) {
	var body ingredientsCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	ingredient := repository.IngredientModel{
		Name:          body.Name,
		Count:         body.Count,
		PurchasePrice: body.PurchasePrice,
		MeasureUnit:   body.MeasureUnit,
		OutletID:      claims.OutletID,
		OrgID:         claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			ingredient.OutletID = stdQuery.OutletID
		}
	}

	if err := s.repo.Ingredients.Create(&ingredient); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: ingredient.ID})
}

type IngredientGetAllResponse []IngredientOutputModel

// @Summary Получить все ингредиенты точки
// @Accept json
// @Produce json
// @Success 200 {object} IngredientGetAllResponse "возвращает все ингредиенты текущей точки"
// @Failure 400 {object} serviceError
// @Failure 500 {object} serviceError
// @Router /ingredients [get]
func (s *ingredients) GetAll(c *gin.Context) {
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
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	var output IngredientGetAllResponse = make(IngredientGetAllResponse, len(*ingredients))
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

type IngredientUpdateInput struct {
	Name          *string  `json:"name,omitempty"`
	Count         *float64 `json:"count,omitempty"`
	PurchasePrice *float64 `json:"purchase_price,omitempty"`
	MeasureUnit   *int     `json:"measure_unit,omitempty"`
}

// @Summary Обновить ингредиент
// @param type body IngredientUpdateInput false "Обновляемые поля"
// @Success 200 {object} object "возвращает пустой объект"
// @Accept json
// @Produce json
// @Failure 400 {object} serviceError
// @Failure 500 {object} serviceError
// @Router /ingredients [put]
func (s *ingredients) UpdateFields(c *gin.Context) {
	var input IngredientUpdateInput
	if err := c.BindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	idx, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.IngredientModel{
		ID:       uint(idx),
		OutletID: claims.OutletID,
		OrgID:    claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	updated := make(map[string]interface{})
	{
		if input.Name != nil {
			if *input.Name != "" {
				updated["name"] = *input.Name
			}
		}

		if input.PurchasePrice != nil {
			updated["purchase_price"] = *input.PurchasePrice
		}

		if input.Count != nil {
			updated["count"] = *input.Count
		}

		if input.MeasureUnit != nil {
			if *input.MeasureUnit < 1 || *input.MeasureUnit > 3 {
				NewResponse(c, http.StatusBadRequest, errIncorrectInputData("1 <= measure_unit <= 3"))
				return
			}
			updated["measure_unit"] = *input.MeasureUnit
		}

	}

	if err := s.repo.Ingredients.UpdatesFull(where, &updated); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

// @Summary Удаляет ингридиент из точки
// @Accept json
// @Produce json
// @Success 201 {object} object "возвращает пустой объект"
// @Failure 400 {object} serviceError
// @Failure 500 {object} serviceError
// @Router /ingredients/:id [delete]
func (s *ingredients) Delete(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)
	stdQuery := mustGetStdQuery(c)

	ingrID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	where := &repository.IngredientModel{
		ID:       uint(ingrID),
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	if err := s.repo.Ingredients.Delete(where); err != nil {
		if dberr, ok := isDatabaseError(err); ok {
			switch dberr.Number {
			case 1451:
				NewResponse(c, http.StatusBadRequest, errForeignKey("the ingredient has not deleted communications"))
				return
			}
		}
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

type IngredientArrivalInput struct {
	IngredientID uint    `json:"ingredient_id" binding:"min=1"`
	Count        float64 `json:"count" binding:"min=0"`
	WriteOff     bool    `json:"write_off"`
	Price        float64 `json:"price" binding:"min=0"`
	Date         int64   `json:"date" binding:"min=1"`
}

// @Summary Поступление ингредиентов в точку
// @param type body []IngredientArrivalInput false "Принимаемый объект"
// @Accept json
// @Produce json
// @Success 201 {object} object "возвращает пустой объект
// @Failure 400 {object} serviceError
// @Failure 500 {object} serviceError
// @Router /ingredients.Arrival [post]
func (s *ingredients) Arrival(c *gin.Context) {
	var input []IngredientArrivalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	if len(input) > 100 {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("you can't transfer more than 100 objects"))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	ingredients := make([]*repository.IngredientModel, len(input))

	where := &repository.IngredientModel{
		OutletID: claims.OutletID,
		OrgID:    claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			where.OutletID = claims.OutletID
		}
	}

	//получение инфы об ингредиентах
	for i, arrival := range input {
		where.ID = arrival.IngredientID
		ingredient, err := s.repo.Ingredients.FindFirts(where)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				NewResponse(c, http.StatusBadRequest, errRecordNotFound(fmt.Sprintf("undefined ingredent with id `%d`", where.ID)))
			} else {
				NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
			}
			return
		}
		ingredients[i] = ingredient
	}

	//обновление и добавление в историю (ingredientsAddingHistory)
	var writeOffSum float64
	for i, arrival := range input {
		if arrival.WriteOff {
			writeOffSum += arrival.Price * arrival.Count
		}

		ingredients[i].Count += arrival.Count
		s.repo.Ingredients.Updates(&repository.IngredientModel{ID: arrival.IngredientID}, ingredients[i])

		//добавление в историю
		if err := s.repo.IngredientsAddingHistory.Create(&repository.IngredientsAddingHistoryModel{
			Count:  arrival.Count,
			Total:  arrival.Count * arrival.Price,
			Status: 3,
			Date:   arrival.Date,

			IngredientID: arrival.IngredientID,
			EmployeeID:   claims.EmployeeID,
			OutletID:     claims.OutletID,
			OrgID:        claims.OrganizationID,
		}); err != nil {
			NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		}
	}

	//добавление инфы в кассу
	model := &repository.CashChangesModel{
		Date:       time.Now().UTC().UnixMilli(),
		Total:      writeOffSum,
		Reason:     "receipt of goods",
		EmployeeID: claims.EmployeeID,
		OutletID:   claims.OutletID,
		OrgID:      claims.OrganizationID,
	}

	if err := s.repo.CashChanges.Create(model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, nil)
}
