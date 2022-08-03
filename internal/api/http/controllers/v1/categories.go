package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type categories struct {
	repo *repository.Repository
}

type CategoryOutputModel struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	OutletID uint   `json:"outlet_id"`
}

func newCategories(repo *repository.Repository) *categories {
	return &categories{
		repo: repo,
	}
}

type categoriesCreateBody struct {
	Name string `json:"name" binding:"required,max=150"`
}

//@Summary Добавить новую категорию к точке
//@param type body categoriesCreateBody false "Принимаемый объект"
//@Accept json
//@Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
//@Router /categories [post]
func (s *categories) Create(c *gin.Context) {
	var body categoriesCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	categoryModel := repository.CategoryModel{
		Name:     body.Name,
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			categoryModel.OutletID = stdQuery.OutletID
		}
	}

	if err := s.repo.Categories.Create(&categoryModel); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: categoryModel.ID})
}

type categoriesGetAllResponse []CategoryOutputModel

//@Summary Список всех категорий организации для владельца и точки для админа/кассира
//@Produce json
//@Success 200 {object} categoriesGetAllResponse "Возвращает массив категорий"
//@Failure 500 {object} serviceError
//@Router /categories [get]
func (s *categories) GetAll(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)
	stdQuery := mustGetStdQuery(c)

	where := &repository.CategoryModel{
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

	cats, err := s.repo.Categories.Find(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	var output categoriesGetAllResponse = make(categoriesGetAllResponse, len(*cats))
	for i, cat := range *cats {
		output[i] = CategoryOutputModel{
			ID:       cat.ID,
			Name:     cat.Name,
			OutletID: cat.OutletID,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

//@Summary Удалить категоирю из точки
//@Description Метод позволяет удалить категоирю из точки
//@Produce json
//@Success 200 {object} object "возвращает пустой объект"
//@Failure 500 {object} serviceError
//@Router /categories/:id [delete]
func (s *categories) Delete(c *gin.Context) {
	catID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)
	stdQuery := mustGetStdQuery(c)

	where := &repository.CategoryModel{ID: uint(catID), OrgID: claims.OrganizationID, OutletID: claims.OutletID}
	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	if err := s.repo.Categories.Delete(where); err != nil {
		if dberr, ok := isDatabaseError(err); ok {
			switch dberr.Number {
			case 1451:
				NewResponse(c, http.StatusBadRequest, errForeignKey("the category has not deleted products"))
				return
			}
		}
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

type CategoryUpdateFieldsInput struct {
	Name string `json:"name"`
}

//@Summary Обновить поля категории
//@param type body CategoryUpdateFieldsInput false "Принимаемый объект"
//@Accept json
//@Produce json
//@Success 200 {object} object "возвращает пустой объект"
//@Router /categories/:id [put]
func (s *categories) UpdateFields(c *gin.Context) {
	var input CategoryUpdateFieldsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	catID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)
	stdQuery := mustGetStdQuery(c)

	where := &repository.CategoryModel{
		ID:       uint(catID),
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	updatedFields := &repository.CategoryModel{
		Name: input.Name,
	}

	if err := s.repo.Categories.Updates(where, updatedFields); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}
