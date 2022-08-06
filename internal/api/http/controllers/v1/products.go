package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/s3cloud"
	"gorm.io/gorm"
)

type productResponseModel struct {
	ID         uint `json:"id" mapstructure:"id"`
	CategoryID uint `json:"category_id" mapstructure:"category_id"`
	OutletID   uint `json:"outlet_id" mapstructure:"outlet_id"`

	Name           string `json:"name" mapstructure:"name"`
	ProductNameKKT string `json:"product_name_kkt" mapstructure:"product_name_kkt"`

	Amount  int `json:"amount" mapstructure:"amount"`
	Barcode int `json:"barcode" mapstructure:"barcode"`

	Price         float64 `json:"price" mapstructure:"price"`
	SellerPercent float64 `json:"seller_percent" mapstructure:"seller_percent"`

	Photo string `json:"photo" mapstructure:"photo"`
}

type products struct {
	repo    *repository.Repository
	s3cloud *s3cloud.SelectelS3Cloud
}

func newProducts(
	repo *repository.Repository,
	s3cloud *s3cloud.SelectelS3Cloud,
) *products {
	return &products{
		repo:    repo,
		s3cloud: s3cloud,
	}
}

type productsCreateBody struct {
	CategoryID uint   `json:"category_id"`
	PhotoID    string `json:"photo_id" binding:"max=500"`

	Name           string `json:"name" binding:"min=1,max=200"`
	ProductNameKKT string `json:"product_name_kkt" binding:"max=200"`

	Barcode int `json:"barcode" binding:"min=0"`
	Amount  int `json:"amount" binding:"min=0"`

	Price         float64 `json:"price" binding:"min=0"`
	SellerPercent float64 `json:"seller_percent" binding:"min=0,max=100"`
}

// @Summary Добавить новый продукт в точку
// @param type body ProductCreateInput false "Принимаемый объект"
// @Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
// @Accept json
// @Produce json
// @Failure 400 {object} serviceError
// @Failure 500 {object} serviceError
// @Router /products [post]
func (s *products) Create(c *gin.Context) {
	var body productsCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	newProduct := repository.ProductModel{
		CategoryID: body.CategoryID,
		OutletID:   claims.OutletID,
		OrgID:      claims.OrganizationID,

		Name:           body.Name,
		ProductNameKKT: body.ProductNameKKT,

		Barcode: body.Barcode,
		Amount:  body.Amount,

		Price:         body.Price,
		SellerPercent: body.SellerPercent / 100,

		PhotoCloudID: body.PhotoID,
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
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}
	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: newProduct.ID})
}

type productGetAllResponse []productResponseModel

// @Summary Список продуктов точки
// @Success 200 {object} productGetAllResponse "возвращает список пордуктов точки"
// @Accept json
// @Produce json
// @Failure 400 {object} serviceError
// @Failure 500 {object} serviceError
// @Router /products [get]
func (s *products) GetAll(c *gin.Context) {
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
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := make(productGetAllResponse, len(*products))
	for i, product := range *products {
		output[i] = productResponseModel{
			ID: product.ID,

			Name:           product.Name,
			ProductNameKKT: product.ProductNameKKT,

			Barcode:       product.Barcode,
			Amount:        product.Amount,
			Price:         product.Price,
			SellerPercent: product.SellerPercent * 100,

			Photo: s.s3cloud.GetURIFromFileID(product.PhotoCloudID),

			CategoryID: product.CategoryID,
			OutletID:   product.OutletID,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

// @Summary Продукт точки
// @Success 200 {object} ProductOutputModel "возвращает один продукт из точки"
// @Accept json
// @Produce json
// @Failure 400 {object} serviceError
// @Failure 500 {object} serviceError
// @Router /products/:id [get]
func (s *products) GetOne(c *gin.Context) {
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

	output := productResponseModel{
		ID:             product.ID,
		Name:           product.Name,
		ProductNameKKT: product.ProductNameKKT,

		Barcode:       product.Barcode,
		Amount:        product.Amount,
		Price:         product.Price,
		SellerPercent: product.SellerPercent * 100,

		Photo: s.s3cloud.GetURIFromFileID(product.PhotoCloudID),

		CategoryID: product.CategoryID,
		OutletID:   product.OutletID,
	}
	NewResponse(c, http.StatusOK, output)
}

type productUpdateBody struct {
	Name           *string  `json:"name,omitempty"`
	ProductNameKKT *string  `json:"product_name_kkt,omitempty"`
	Barcode        *int     `json:"barcode,omitempty"`
	Amount         *int     `json:"amount,omitempty"`
	Price          *float64 `json:"price,omitempty"`
	SellerPercent  *float64 `json:"seller_percent,omitempty"`
	PhotoID        *string  `json:"photo_id,omitempty"`
	CategoryID     *uint    `json:"category_id,omitempty"`
}

// @Summary Обновить продукт в точке
// @param type body productUpdateBody false "Обновляемые поля"
// @Success 200 {object} object "возвращает пустой объект"
// @Accept json
// @Produce json
// @Failure 400 {object} serviceError
// @Router /products/:id [put]
func (s *products) Update(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	var body productUpdateBody
	if err := c.BindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.ProductModel{
		ID:       uint(productID),
		OutletID: claims.OutletID,
		OrgID:    claims.OrganizationID,
	}

	updated := make(map[string]interface{})
	{
		if body.Name != nil {
			updated["name"] = *body.Name
		}

		if body.ProductNameKKT != nil {
			updated["product_name_kkt"] = *body.ProductNameKKT
		}

		if body.Barcode != nil {
			updated["barcode"] = *body.Barcode
		}

		if body.Amount != nil {
			updated["amount"] = *body.Amount
		}

		if body.Price != nil {
			updated["price"] = *body.Price
		}

		if body.PhotoID != nil {
			updated["photo_cloud_id"] = *body.PhotoID
		}

		if body.SellerPercent != nil {
			if *body.SellerPercent < 0 || *body.SellerPercent > 100 {
				NewResponse(c, http.StatusBadRequest, errIncorrectInputData("0 <= seller_percent <= 100"))
				return
			}
			updated["seller_percent"] = *body.SellerPercent / 100
		}

		if body.CategoryID != nil {
			outletID := claims.OutletID

			if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
				outletID = stdQuery.OutletID
			}

			if !s.repo.Categories.Exists(&repository.CategoryModel{ID: *body.CategoryID, OutletID: outletID}) {
				NewResponse(c, http.StatusBadRequest, errIncorrectInputData("incorrect `category_id`"))
				return
			}

			updated["category_id"] = *body.CategoryID
		}
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	if err := s.repo.Products.UpdatesFull(where, &updated); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

// @Summary Удалить продукт в точке
// @Success 200 {object} object "возвращает пустой объект"
// @Accept json
// @Produce json
// @Failure 400 {object} serviceError
// @Failure 500 {object} serviceError
// @Router /products/:id [delete]
func (s *products) Delete(c *gin.Context) {
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
