package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type productWithIngredientsResponseModel struct {
	ID           uint `json:"id" mapstructure:"id"`
	IngredientID uint `json:"ingredient_id" mapstructure:"ingredient_id"`
	ProductID    uint `json:"product_id" mapstructure:"product_id"`
	OutletID     uint `json:"outlet_id" mapstructure:"outlet_id"`

	CountTakeForSell float64 `json:"count_take_for_sell" mapstructure:"count_take_for_sell"`
}

type productsWithIngredients struct {
	repo *repository.Repository
}

func newProductsWithIngredients(repo *repository.Repository) *productsWithIngredients {
	return &productsWithIngredients{
		repo: repo,
	}
}

type productsWithIngredientsCreateBody struct {
	CountTakeForSell float64 `json:"count_take_for_sell"`
	ProductID        uint    `json:"product_id" binding:"min=1"`
	IngredientID     uint    `json:"ingredient_id" binding:"min=1"`
}

// @Summary Добавить связь продукта и ингридиента в точку
// @param type body productsWithIngredientsCreateBody false "object"
// @Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
// @Router /pwis [post]
func (s *productsWithIngredients) Create(c *gin.Context) {
	var input productsWithIngredientsCreateBody
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)
	stdQuery := mustGetStdQuery(c)

	pwiModel := &repository.ProductWithIngredientModel{
		CountTakeForSell: input.CountTakeForSell,
		IngredientID:     input.IngredientID,
		ProductID:        input.ProductID,
		OutletID:         claims.OutletID,
		OrgID:            claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			pwiModel.OutletID = stdQuery.OutletID
		}
	}

	if !s.repo.Products.Exists(&repository.ProductModel{ID: pwiModel.ProductID, OutletID: pwiModel.OutletID}) ||
		!s.repo.Ingredients.Exists(&repository.IngredientModel{ID: pwiModel.IngredientID, OutletID: pwiModel.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("not found product or ingredient with this `id` in outlet"))
		return
	}

	if err := s.repo.ProductsWithIngredients.Create(pwiModel); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: pwiModel.ID})
}

type productsWithIngredientsGetAllQuery struct {
	ProductID uint `form:"product_id"`
}

type productsWithIngredientsGetAllResponse []productWithIngredientsResponseModel

// @Summary Получить список связей продуктов и ингредиентов в точке
// @Param type query productsWithIngredientsGetAllQuery false "query"
// @Success 200 {object} productsWithIngredientsGetAllResponse "список связей мужду продуктами и ингредиентами"
// @Router /pwis [get]
func (s *productsWithIngredients) GetAll(c *gin.Context) {
	var query productsWithIngredientsGetAllQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.ProductWithIngredientModel{
		ProductID: query.ProductID,
		OutletID:  claims.OutletID,
		OrgID:     claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER) {
		if stdQuery.OrgID != 0 && s.repo.Invitation.Exists(&repository.InvitationModel{OrgID: claims.OrganizationID, AffiliateOrgID: stdQuery.OrgID}) {
			where.OrgID = stdQuery.OrgID
		}
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	pwis, err := s.repo.ProductsWithIngredients.Find(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := make(productsWithIngredientsGetAllResponse, len(*pwis))
	for i, pwi := range *pwis {
		output[i] = productWithIngredientsResponseModel{
			ID:               pwi.ID,
			CountTakeForSell: pwi.CountTakeForSell,
			ProductID:        pwi.ProductID,
			IngredientID:     pwi.IngredientID,
			OutletID:         pwi.OutletID,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

// @Summary Удалить связь из точки
// @Success 200 {object} object "object"
// @Router /pwis/:id [delete]
func (s *productsWithIngredients) Delete(c *gin.Context) {
	pwiID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.ProductWithIngredientModel{ID: uint(pwiID), OrgID: claims.OrganizationID, OutletID: claims.OutletID}
	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	if err := s.repo.ProductsWithIngredients.Delete(where); err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

type productsWithIngredientsUpdate struct {
	ProductID        uint    `json:"product_id"`
	CountTakeForSell float64 `json:"count_take_for_sell"`
}

// @Summary Обновить связь
// @Param type body PWIUpdateFields false "Обновляемые поля"
// @Success 200 {object} object "возвращает пустой объект"
// @Router /pwis/:id [put]
func (s *productsWithIngredients) Update(c *gin.Context) {
	var input productsWithIngredientsUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	pwiID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.ProductWithIngredientModel{ID: uint(pwiID), OrgID: claims.OrganizationID, OutletID: claims.OutletID}
	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	updatedFields := &repository.ProductWithIngredientModel{
		CountTakeForSell: input.CountTakeForSell,
		ProductID:        input.ProductID,
	}

	if err := s.repo.ProductsWithIngredients.Updates(where, updatedFields); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}
