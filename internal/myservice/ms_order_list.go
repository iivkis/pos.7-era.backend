package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
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

	if err := s.repo.ProductsWithIngredients.WriteOffIngredients(input.ProductID, input.Count, c.MustGet("claims_outlet_id").(uint)); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	//TODO: проверка есть ли orderInfo в точке
	m := repository.OrderListModel{
		ProductName:  input.ProductName,
		ProductPrice: input.ProductPrice,
		ProductID:    input.ProductID,
		Count:        input.Count,
		OrderInfoID:  input.OrderInfoID,
		OutletID:     c.MustGet("claims_outlet_id").(uint),
		OrgID:        c.MustGet("claims_org_id").(uint),
	}

	if err := s.repo.OrdersList.Create(&m); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: m.ID})
}

type OrderListGetAllForOutletOutput []OrderListOutputModel

//@Summary Получить список orderList точки (список продутктов из которых состоит заказ)
//@Success 200 {object} OrderListGetAllForOutletOutput "список orderList точки"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /orderList [get]
func (s *OrdersListService) GetAllForOutlet(c *gin.Context) {
	models, err := s.repo.OrdersList.FindAllByOutletID(c.MustGet("claims_outlet_id").(uint))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
	}

	var output OrderListGetAllForOutletOutput = make(OrderListGetAllForOutletOutput, len(models))
	for i, item := range models {
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
