package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type orderListResponseModel struct {
	ID          uint `json:"id" mapstructure:"id"`
	ProductID   uint `json:"product_id" mapstructure:"product_id"`
	OrderInfoID uint `json:"order_info_id" mapstructure:"order_info_id"`
	SessionID   uint `json:"session_id" mapstructure:"session_id"`
	OutletID    uint `json:"outlet_id" mapstructure:"outlet_id"`

	Count        int     `json:"count" mapstructure:"count"`
	ProductName  string  `json:"product_name" mapstructure:"product_name"`
	ProductPrice float64 `json:"product_price" mapstructure:"product_price"`
}

type orderList struct {
	repo *repository.Repository
}

func newOrderList(repo *repository.Repository) *orderList {
	return &orderList{
		repo: repo,
	}
}

type orderListCreateBody struct {
	ProductID   uint `json:"product_id" binding:"min=1"`
	OrderInfoID uint `json:"order_info_id" binding:"min=1"`
	SessionID   uint `json:"session_id" binding:"min=1"`

	Count        int     `json:"count"`
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`
}

// @Summary Добавить продукт в чек (в orderInfo)
// @param type body orderListCreateBody false "object"
// @Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
// @Router /orderList [post]
func (s *orderList) Create(c *gin.Context) {
	var input orderListCreateBody
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	ordersList := repository.OrderListModel{
		ProductName:  input.ProductName,
		ProductPrice: input.ProductPrice,
		ProductID:    input.ProductID,
		Count:        input.Count,
		OrderInfoID:  input.OrderInfoID,
		SessionID:    input.SessionID,
		OutletID:     claims.OutletID,
		OrgID:        claims.OrganizationID,
	}

	if !s.repo.Sessions.Exists(&repository.SessionModel{Model: gorm.Model{ID: ordersList.SessionID}, OutletID: ordersList.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `session_id` with this `id`"))
		return
	}

	if !s.repo.OrdersInfo.Exists(&repository.OrderInfoModel{Model: gorm.Model{ID: ordersList.OrderInfoID}, OutletID: ordersList.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `order_info_id` with this `id`"))
		return
	}

	if !s.repo.Products.Exists(&repository.ProductModel{ID: ordersList.ProductID, OutletID: ordersList.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `product_id` with this `id`"))
		return
	}

	if err := s.repo.ProductsWithIngredients.SubractionIngredients(ordersList.ProductID, ordersList.Count); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	if err := s.repo.OrdersList.Create(&ordersList); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: ordersList.ID})
}

type orderListGetAllQuery struct {
	ProductID   uint `form:"product_id"`
	SessionID   uint `form:"session_id"`
	OrderInfoID uint `form:"order_info_id"`
}

type orderListGetAllResponse []orderListResponseModel

// @Summary Получить список orderList точки (список продутктов из которых состоит заказ)
// @Param type query orderListGetAllQuery false "Принимаемый объект"
// @Success 200 {object} orderListGetAllResponse "список orderList точки"
// @Failure 400 {object} serviceError
// @Router /orderList [get]
func (s *orderList) GetAll(c *gin.Context) {
	var query orderListGetAllQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.OrderListModel{
		OrgID:       claims.OrganizationID,
		OutletID:    claims.OutletID,
		OrderInfoID: query.OrderInfoID,
		SessionID:   query.SessionID,
		ProductID:   query.ProductID,
	}

	if claims.HasRole(repository.R_OWNER) {
		if stdQuery.OrgID != 0 && s.repo.Invitation.Exists(&repository.InvitationModel{OrgID: claims.OrganizationID, AffiliateOrgID: stdQuery.OrgID}) {
			where.OrgID = stdQuery.OrgID
		}
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	models, err := s.repo.OrdersList.FindUnscoped(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
	}

	var output orderListGetAllResponse = make(orderListGetAllResponse, len(*models))
	for i, item := range *models {
		output[i] = orderListResponseModel{
			ID:           item.ID,
			Count:        item.Count,
			ProductName:  item.ProductName,
			ProductPrice: item.ProductPrice,
			ProductID:    item.ProductID,
			OrderInfoID:  item.OrderInfoID,
			SessionID:    item.SessionID,
			OutletID:     item.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}

type orderListCalcResponse struct {
	Total float64 `json:"total" mapstructure:"total"`
}

// @Summary  Посчитать сумму продаж за определенный период
// @Param type query OrderListGetAllQuery false "Принимаемый объект"
// @Accept json
// @Produce json
// @Success 200 {object} OrderListCalcOutput "сумма с продаж"
// @Failure 400 {object} serviceError
// @Router /orderList.Calc [get]
func (s *orderList) Calc(c *gin.Context) {
	var query orderListGetAllQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.OrderListModel{
		OrgID:       claims.OrganizationID,
		OutletID:    claims.OutletID,
		OrderInfoID: query.OrderInfoID,
		SessionID:   query.SessionID,
		ProductID:   query.ProductID,
	}

	if claims.HasRole(repository.R_OWNER) {
		if stdQuery.OrgID != 0 && s.repo.Invitation.Exists(&repository.InvitationModel{
			OrgID:          claims.OrganizationID,
			AffiliateOrgID: stdQuery.OrgID,
		}) {
			where.OrgID = stdQuery.OrgID
		}
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	list, err := s.repo.OrdersList.FindForCalculation(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
	}

	response := orderListCalcResponse{}
	for _, item := range *list {
		response.Total += item.ProductPrice * float64(item.Count)
	}

	NewResponse(c, http.StatusOK, response)
}
