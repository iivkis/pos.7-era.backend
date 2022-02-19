package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

type CategoriesService struct {
	repo *repository.Repository
}

type CategoryOutputModel struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	OutletID uint   `json:"outlet_id"`
}

func newCategoriesService(repo *repository.Repository) *CategoriesService {
	return &CategoriesService{
		repo: repo,
	}
}

type CategoryCreateInput struct {
	Name string `json:"name" binding:"required,max=150"`
}

//@Summary Добавить новую категорию к точке
//@param type body CategoryCreateInput false "Принимаемый объект"
//@Accept json
//@Success 201 {object} object "возвращает пустой объект"
//@Router /categories [post]
func (s *CategoriesService) Create(c *gin.Context) {
	var input CategoryCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	cat := repository.CategoryModel{
		Name:     input.Name,
		OrgID:    c.MustGet("claims_org_id").(uint),
		OutletID: c.MustGet("claims_outlet_id").(uint),
	}
	if err := s.repo.Categories.Create(&cat); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, nil)
}

type CategoryGetAllOutput []CategoryOutputModel

//@Summary Список всех категорий точки
//@Description Метод позволяет получить список категорий точки
//@Produce json
//@Success 200 {object} CategoryGetAllOutput "Возвращает массив категорий"
//@Failure 500 {object} serviceError
//@Router /categories [get]
func (s *CategoriesService) GetAll(c *gin.Context) {
	cats, err := s.repo.Categories.GetAllByOrgID(c.MustGet("claims_org_id").(uint))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output CategoryGetAllOutput = make(CategoryGetAllOutput, len(cats))
	for i, cat := range cats {
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
func (s *CategoriesService) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := s.repo.Categories.DeleteByID(c.MustGet("claims_outlet_id").(uint), id); err != nil {
		NewResponse(c, http.StatusBadRequest, errOnDelet(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

type UpdateCategoryInput struct {
	Name string `json:"name"`
}

//@Summary Обновить поля категории
//@param type body UpdateCategoryInput false "Принимаемый объект"
//@Accept json
//@Produce json
//@Success 201 {object} object "возвращает пустой объект"
//@Router /categories/:id [put]
func (s *CategoriesService) Update(c *gin.Context) {
	var input UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	cat := repository.CategoryModel{
		Name: input.Name,
	}

	if err := s.repo.Categories.Updates(c.Param("id"), c.MustGet("claims_outlet_id"), &cat); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}
