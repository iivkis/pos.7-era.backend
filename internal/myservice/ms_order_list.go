package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type OrderListOutputModel struct {
	ID           uint    `json:"id"`
	Count        int     `json:"count"`
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`

	ProductID   uint `json:"product_id"`
	OrderInfoID uint `json:"order_info_id"`
	OutletID    uint `json:"outlet_id"`
}

type OrdersListService struct {
	repo *repository.Repository
}

func newOrderListService(repo *repository.Repository) *OrdersListService {
	return &OrdersListService{
		repo: repo,
	}
}

type OrderListCreateInput struct {
	Count        int     `json:"count"`
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`

	ProductID   uint `json:"product_id"`
	OrderInfoID uint `json:"order_info_id"`
}

//@Summary Добавить orderList (список продутктов из которых состоит заказ)
//@param type body OrderListCreateInput false "Принимаемый объект"
//@Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /orderList [post]
func (s *OrdersListService) Create(c *gin.Context) {
	var input OrderListCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	model := repository.OrderListModel{
		ProductName:  input.ProductName,
		ProductPrice: input.ProductPrice,
		ProductID:    input.ProductID,
		Count:        input.Count,
		OrderInfoID:  input.OrderInfoID,
		OutletID:     claims.OutletID,
		OrgID:        claims.OrganizationID,
	}

	if model.OrderInfoID == 0 || !s.repo.OrdersInfo.Exists(&repository.OrderInfoModel{Model: gorm.Model{ID: model.OrderInfoID}, OutletID: model.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `order_info_id` with this `id`"))
		return
	}

	if model.ProductID == 0 || !s.repo.Products.Exists(&repository.ProductModel{Model: gorm.Model{ID: model.ProductID}, OutletID: model.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `product_id` with this `id`"))
		return
	}

	if err := s.repo.ProductsWithIngredients.WriteOffIngredients(model.ProductID, model.Count); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	if err := s.repo.OrdersList.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type OrderListGetAllOutput []OrderListOutputModel

//@Summary Получить список orderList точки (список продутктов из которых состоит заказ)
//@Success 200 {object} OrderListGetAllOutput "список orderList точки"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /orderList [get]
func (s *OrdersListService) GetAll(c *gin.Context) {
	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.OrderListModel{
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	models, err := s.repo.OrdersList.Find(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
	}

	var output OrderListGetAllOutput = make(OrderListGetAllOutput, len(*models))
	for i, item := range *models {
		output[i] = OrderListOutputModel{
			ID:           item.ID,
			Count:        item.Count,
			ProductName:  item.ProductName,
			ProductPrice: item.ProductPrice,
			ProductID:    item.ProductID,
			OrderInfoID:  item.OrderInfoID,
			OutletID:     item.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}
