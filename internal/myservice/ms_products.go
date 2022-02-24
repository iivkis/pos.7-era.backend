package myservice

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"gorm.io/gorm"
)

type ProductOutputModel struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Amount     int     `json:"amount"`
	Price      float64 `json:"price"`
	Photo      string  `json:"photo"`
	CategoryID uint    `json:"category_id"`
	OutletID   uint    `json:"outlet_id"`
}

type ProductsService struct {
	repo *repository.Repository
}

func newProductsService(repo *repository.Repository) *ProductsService {
	return &ProductsService{
		repo: repo,
	}
}

type ProductCreateInput struct {
	Name       string  `json:"name" binding:"min=1"`
	Amount     int     `json:"amount"`
	Price      float64 `json:"price"`
	Photo      string  `json:"photo"`
	CategoryID uint    `json:"category_id" binding:"min=1"`
}

//@Summary Добавить новый продукт в точку
//@param type body ProductCreateInput false "Принимаемый объект"
//@Accept json
//@Success 201 {object} object "возвращает пустой объект"
//@Router /products [post]
func (s *ProductsService) Create(c *gin.Context) {
	var input ProductCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
	}

	newProduct := repository.ProductModel{
		Name:       input.Name,
		Amount:     input.Amount,
		Price:      input.Price,
		Photo:      input.Photo,
		CategoryID: input.CategoryID,
		OutletID:   c.MustGet("claims_outlet_id").(uint),
		OrgID:      c.MustGet("claims_org_id").(uint),
	}

	if err := s.repo.Products.Create(&newProduct); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errIncorrectInputData("incorrect `category_id`"))
			return
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusCreated, nil)
}

type ProductGetAllForOrgOutput []ProductOutputModel

//@Summary Список продуктов организации
//@Accept json
//@Success 201 {object} ProductGetAllForOrgOutput "возвращает список пордуктов организации"
//@Router /products [get]
func (s *ProductsService) GetAllForOrg(c *gin.Context) {
	products, err := s.repo.Products.FindAllByOrgID(c.MustGet("claims_org_id"))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := make(ProductGetAllForOrgOutput, len(products))
	for i, product := range products {
		output[i] = ProductOutputModel{
			ID:         product.ID,
			Name:       product.Name,
			Amount:     product.Amount,
			Price:      product.Price,
			Photo:      product.Photo,
			CategoryID: product.CategoryID,
			OutletID:   product.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}

type ProductGetAllForOutletOutput []ProductOutputModel

//@Summary Список продуктов точки
//@Accept json
//@Success 201 {object} ProductGetAllForOutletOutput "возвращает список пордуктов точки"
//@Router /products.Outlet [get]
func (s *ProductsService) GetAllForOutlet(c *gin.Context) {
	products, err := s.repo.Products.FindAllByOutletID(c.MustGet("claims_outlet_id"))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := make(ProductGetAllForOutletOutput, len(products))
	for i, product := range products {
		output[i] = ProductOutputModel{
			ID:         product.ID,
			Name:       product.Name,
			Amount:     product.Amount,
			Price:      product.Price,
			Photo:      product.Photo,
			CategoryID: product.CategoryID,
			OutletID:   product.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}

//@Summary Продукт организации
//@Accept json
//@Success 201 {object} ProductGetAllForOutletOutput "возвращает список пордуктов организации"
//@Router /products/:id [get]
func (s *ProductsService) GetOneForOutlet(c *gin.Context) {
	product, err := s.repo.Products.FindOneByOutletID(c.Param("id"), c.MustGet("claims_outlet_id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound())
			return
		}
		NewResponse(c, http.StatusInternalServerError, errRecordNotFound())
		return
	}

	output := ProductOutputModel{
		ID:         product.ID,
		Name:       product.Name,
		Amount:     product.Amount,
		Price:      product.Price,
		Photo:      product.Photo,
		CategoryID: product.CategoryID,
		OutletID:   product.OutletID,
	}
	NewResponse(c, http.StatusOK, output)
}

type ProductUpdateInput struct {
	Name   string  `json:"name"`
	Amount int     `json:"amount"`
	Price  float64 `json:"price"`
	Photo  string  `json:"photo"`
}

//@Summary Обновить продукт
//@param type body ProductUpdateInput false "Обновляемые поля"
//@Accept json
//@Success 200 {object} object "возвращает пустой объект"
//@Router /products [put]
func (s *ProductsService) UpdateFields(c *gin.Context) {
	var input ProductUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	product := repository.ProductModel{
		Name:   input.Name,
		Amount: input.Amount,
		Price:  input.Price,
		Photo:  input.Photo,
	}

	if err := s.repo.Products.Updates(c.Param("id"), c.MustGet("claims_outlet_id"), &product); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

//@Summary Удалить продукт
//@Accept json
//@Success 200 {object} object "возвращает пустой объект"
//@Router /products/:id [delete]
func (s *ProductsService) Delete(c *gin.Context) {
	if err := s.repo.Products.Delete(c.Param("id"), c.MustGet("claims_outlet_id")); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}
