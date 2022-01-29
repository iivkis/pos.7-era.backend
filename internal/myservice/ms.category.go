package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

type CategoryService interface {
	Create(*gin.Context)
}

type category struct {
	repo repository.Repository
}

func newCategoryService(repo repository.Repository) *category {
	return &category{
		repo: repo,
	}
}

type createCategoryInput struct {
	Name string `json:"name" binding:"max=150"`
}

func (s *category) Create(c *gin.Context) {
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
