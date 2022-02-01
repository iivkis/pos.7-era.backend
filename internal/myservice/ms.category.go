package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

type CategoriesService struct {
	repo repository.Repository
}

type categoryOutputModel struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func newCategoriesService(repo repository.Repository) *CategoriesService {
	return &CategoriesService{
		repo: repo,
	}
}

type createCategoryInput struct {
	Name string `json:"name" binding:"max=150"`
}

//@Summary Добавить новую категорию к точке
//@param type body createCategoryInput false "Принимаемый объект"
//@Accept json
//@Success 201 {object} object "возвращает пустой объект"
//@Router /category [post]
func (s *CategoriesService) Create(c *gin.Context) {
	var input createCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	cat := repository.CategoryModel{
		Name:     input.Name,
		OutletID: c.MustGet("claims_outlet_id").(uint),
	}
	if err := s.repo.Category.Create(&cat); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, nil)
}

type getAllCategoryOutput []categoryOutputModel

//@Summary Список всех категорий точки
//@Description Метод позволяет получить список категорий точки
//@Produce json
//@Success 200 {object} getAllCategoryOutput "Возвращает массив категорий"
//@Failure 500 {object} serviceError
//@Router /category [get]
func (s *CategoriesService) GetAll(c *gin.Context) {
	cats, err := s.repo.Category.GetAllByOutletID(c.MustGet("claims_outlet_id").(uint))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output getAllCategoryOutput = make(getAllCategoryOutput, len(cats))
	for i, cat := range cats {
		output[i] = categoryOutputModel{
			ID:   cat.ID,
			Name: cat.Name,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

//@Summary Удалить категоирю из точки
//@Description Метод позволяет удалить категоирю из точки
//@Produce json
//@Success 200 {object} object "возвращает пустой объект"
//@Failure 500 {object} serviceError
//@Router /category/:id [delete]
func (s *CategoriesService) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("empty param `id`"))
		return
	}

	if err := s.repo.Category.DeleteByID(c.MustGet("claims_outlet_id").(uint), id); err != nil {
		NewResponse(c, http.StatusBadRequest, errOnDelet(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

//@Summary Обновить поля категории
//@param type body createCategoryInput false "Принимаемый объект"
//@Accept json
//@Produce json
//@Success 201 {object} object "возвращает пустой объект"
//@Router /category/:id [put]
func (s *CategoriesService) Update(c *gin.Context) {
	var input createCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	cat := repository.CategoryModel{
		Name: input.Name,
	}

	if err := s.repo.Category.Update(c.Param("id"), &cat); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}
