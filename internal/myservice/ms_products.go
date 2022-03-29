package myservice

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type ProductOutputModel struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	ProductNameKKT string  `json:"product_name_kkt"`
	Barcode        int     `json:"barcode"`
	Amount         int     `json:"amount"`
	Price          float64 `json:"price"`
	Photo          string  `json:"photo"`
	CategoryID     uint    `json:"category_id"`
	OutletID       uint    `json:"outlet_id"`
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
	Name           string  `json:"name" binding:"min=1,max=200"`
	ProductNameKKT string  `json:"product_name_kkt" binding:"max=200"`
	Barcode        int     `json:"barcode"`
	Amount         int     `json:"amount"`
	Price          float64 `json:"price"`
	Photo          string  `json:"photo"`
	CategoryID     uint    `json:"category_id"`
}

//@Summary Добавить новый продукт в точку
//@param type body ProductCreateInput false "Принимаемый объект"
//@Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /products [post]
func (s *ProductsService) Create(c *gin.Context) {
	var input ProductCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)
	newProduct := repository.ProductModel{
		Name:           input.Name,
		ProductNameKKT: input.ProductNameKKT,
		Barcode:        input.Barcode,
		Amount:         input.Amount,
		Price:          input.Price,
		Photo:          input.Photo,
		CategoryID:     input.CategoryID,
		OutletID:       claims.OutletID,
		OrgID:          claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			newProduct.OutletID = stdQuery.OutletID
		}
	}

	if !s.repo.Categories.Exists(&repository.CategoryModel{ID: newProduct.CategoryID, OutletID: newProduct.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `category` with this id"))
		return
	}

	if err := s.repo.Products.Create(&newProduct); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: newProduct.ID})
}

type ProductGetAllOutput []ProductOutputModel

//@Summary Список продуктов точки
//@Success 200 {object} ProductGetAllOutput "возвращает список пордуктов точки"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /products [get]
func (s *ProductsService) GetAll(c *gin.Context) {
	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.ProductModel{
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

	products, err := s.repo.Products.Find(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := make(ProductGetAllOutput, len(*products))
	for i, product := range *products {
		output[i] = ProductOutputModel{
			ID:             product.ID,
			Name:           product.Name,
			ProductNameKKT: product.ProductNameKKT,
			Barcode:        product.Barcode,
			Amount:         product.Amount,
			Price:          product.Price,
			Photo:          product.Photo,
			CategoryID:     product.CategoryID,
			OutletID:       product.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}

//@Summary Продукт точки
//@Success 200 {object} ProductOutputModel "возвращает один продукт из точки"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /products/:id [get]
func (s *ProductsService) GetOne(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.ProductModel{
		ID:       uint(productID),
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	product, err := s.repo.Products.FindFirst(where)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound())
			return
		}
		NewResponse(c, http.StatusInternalServerError, errRecordNotFound())
		return
	}

	output := ProductOutputModel{
		ID:             product.ID,
		Name:           product.Name,
		ProductNameKKT: product.ProductNameKKT,
		Barcode:        product.Barcode,
		Amount:         product.Amount,
		Price:          product.Price,
		Photo:          product.Photo,
		CategoryID:     product.CategoryID,
		OutletID:       product.OutletID,
	}
	NewResponse(c, http.StatusOK, output)
}

type ProductUpdateInput struct {
	Name           string  `json:"name"`
	ProductNameKKT string  `json:"product_name_kkt"`
	Barcode        int     `json:"barcode"`
	Amount         int     `json:"amount"`
	Price          float64 `json:"price"`
	Photo          string  `json:"photo"`
	CategoryID     uint    `json:"category_id"`
}

//@Summary Обновить продукт в точке
//@param type body ProductUpdateInput false "Обновляемые поля"
//@Success 200 {object} object "возвращает пустой объект"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /products/:id [put]
func (s *ProductsService) UpdateFields(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	var input ProductUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.ProductModel{
		ID:       uint(productID),
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	upadtedFields := &repository.ProductModel{
		Name:           input.Name,
		ProductNameKKT: input.ProductNameKKT,
		Barcode:        input.Barcode,
		Amount:         input.Amount,
		Price:          input.Price,
		Photo:          input.Photo,
		CategoryID:     input.CategoryID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			where.OutletID = stdQuery.OutletID
		}
	}

	if upadtedFields.CategoryID != 0 && !s.repo.Categories.Exists(&repository.CategoryModel{ID: upadtedFields.CategoryID, OutletID: claims.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("incorrect `category_id`"))
		return
	}

	if err := s.repo.Products.Updates(where, upadtedFields); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

//@Summary Удалить продукт в точке
//@Success 200 {object} object "возвращает пустой объект"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /products/:id [delete]
func (s *ProductsService) Delete(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.ProductModel{ID: uint(productID), OrgID: claims.OrganizationID, OutletID: claims.OutletID}
	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	if err := s.repo.Products.Delete(where); err != nil {
		if dberr, ok := isDatabaseError(err); ok {
			switch dberr.Number {
			case 1451:
				NewResponse(c, http.StatusBadRequest, errForeignKey("the product has not deleted communications"))
				return
			}
		}
	}

	NewResponse(c, http.StatusOK, nil)
}
