package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
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

type OrderListCreateOutput struct {
	ID uint `json:"id"`
}

//@Summary Добавить order list
//@param type body OrderListCreateInput false "Принимаемый объект"
//@Accept json
//@Success 201 {object} object "возвращает пустой объект"
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

	newModel := repository.OrderListModel{
		ProductName:  input.ProductName,
		ProductPrice: input.ProductPrice,
		ProductID:    input.ProductID,
		Count:        input.Count,
		OrderInfoID:  input.OrderInfoID,
		OutletID:     c.MustGet("claims_outlet_id").(uint),
		OrgID:        c.MustGet("claims_org_id").(uint),
	}

	if err := s.repo.OrdersList.Create(&newModel); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := OrderListCreateOutput{ID: newModel.ID}
	NewResponse(c, http.StatusCreated, output)
}

type OrderListGetAllForOrgOutput []OrderListOutputModel

//@Summary Получить список order list организации
//@Accept json
//@Success 200 {object} OrderListGetAllForOrgOutput "список order list"
//@Router /orderList [get]
func (s *OrdersListService) GetAllForOrg(c *gin.Context) {
	models, err := s.repo.OrdersList.FindAllForOrg(c.MustGet("claims_org_id").(uint))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
	}

	var output OrderListGetAllForOrgOutput = make(OrderListGetAllForOrgOutput, len(models))
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
