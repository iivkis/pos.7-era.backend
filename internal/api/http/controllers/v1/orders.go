package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type ordersInfoResponse struct {
	SessionID uint `json:"session_id" binding:"min=1"`
	OutletID  uint `json:"outlet_id"`

	EmployeeName string `json:"employee_name" binding:"required"`
	PayType      int    `json:"pay_type" binding:"min=0,max=2"`
	Date         int64  `json:"date" binding:"min=1"` //UnixMilli
	IsDelete     bool   `json:"is_delete"`
}

type ordersListResponse struct {
	ProductID    uint    `json:"product_id" binding:"min=1"`
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`
	Count        int     `json:"count"`
}

type ordersResponseModel struct {
	ID   uint                 `json:"id"` //orderInfo ID
	Info ordersInfoResponse   `json:"info"`
	List []ordersListResponse `json:"list"`
}

type orders struct {
	repo *repository.Repository
}

func newOrders(repo *repository.Repository) *orders {
	return &orders{
		repo: repo,
	}
}

type ordersCreateBody struct {
	Info struct {
		SessionID    uint   `json:"session_id" binding:"min=1"`
		EmployeeName string `json:"employee_name" binding:"required"`
		PayType      int    `json:"pay_type" binding:"min=0,max=2"`
		Date         int64  `json:"date" binding:"min=1"` //UnixMilli
	} `json:"info"`

	List []struct {
		ProductID    uint    `json:"product_id" binding:"min=1"`
		ProductName  string  `json:"product_name"`
		ProductPrice float64 `json:"product_price"`
		Count        int     `json:"count"`
	} `json:"list"`
}

// @Summary Добавить чек
// @Param type body ordersCreateBody false "Принимаемый объект"
// @Success 201 {object} DefaultOutputModel "id созданного чека"
// @Router /orders [post]
func (s *orders) Create(c *gin.Context) {
	var body ordersCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	sess, err := s.repo.Sessions.FindFirts(&repository.SessionModel{Model: gorm.Model{ID: body.Info.SessionID}, OutletID: claims.OutletID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound("undefined session"))
			return
		}
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	if sess.DateClose != 0 {
		NewResponse(c, http.StatusInternalServerError, errIncorrectInputData("session already closed"))
		return
	}

	orderInfoModel := repository.OrderInfoModel{
		SessionID: body.Info.SessionID,
		OrgID:     claims.OrganizationID,
		OutletID:  claims.OutletID,

		PayType:      body.Info.PayType,
		Date:         body.Info.Date,
		EmployeeName: body.Info.EmployeeName,
	}

	if err = s.repo.OrdersInfo.Create(&orderInfoModel); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	for _, item := range body.List {
		orderList := repository.OrderListModel{
			ProductID:   item.ProductID,
			OrderInfoID: orderInfoModel.ID,
			SessionID:   body.Info.SessionID,
			OutletID:    claims.OutletID,
			OrgID:       claims.OrganizationID,

			ProductName:  item.ProductName,
			ProductPrice: item.ProductPrice,
			Count:        item.Count,
		}

		if err := s.repo.ProductsWithIngredients.SubractionIngredients(orderList.ProductID, orderList.Count); err != nil {
			NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
			return
		}

		if err := s.repo.OrdersList.Create(&orderList); err != nil {
			NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
			return
		}
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: orderInfoModel.ID})
}

type ordersGetAllQuery struct {
	SessionID uint `form:"session_id"`
}

type ordersGetAllResponse []ordersResponseModel

// @Summary получить все чеки
// @Param type query orderInfoGetAllQuery false "query"
// @Success 200 {object} orderInfoGetAllResponse "список завершенных заказов"
// @Router /orders [get]
func (s *orders) GetAll(c *gin.Context) {
	var query ordersGetAllQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	whereOrderInfo := &repository.OrderInfoModel{
		OrgID:     claims.OrganizationID,
		OutletID:  claims.OutletID,
		SessionID: query.SessionID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		whereOrderInfo.OutletID = stdQuery.OutletID
	}

	orderInfos, err := s.repo.OrdersInfo.FindUnscoped(whereOrderInfo)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := make(ordersGetAllResponse, len(*orderInfos))

	for i, item := range *orderInfos {
		orderList, _ := s.repo.OrdersList.FindUnscoped(&repository.OrderListModel{
			OrderInfoID: item.ID,
		})

		output[i] = ordersResponseModel{
			ID: item.ID,

			Info: ordersInfoResponse{
				SessionID: item.SessionID,
				OutletID:  item.OutletID,

				EmployeeName: item.EmployeeName,
				PayType:      item.PayType,
				Date:         item.Date,
				IsDelete:     !item.DeletedAt.Time.IsZero(),
			},

			List: make([]ordersListResponse, len(*orderList)),
		}

		for j, item := range *orderList {
			output[i].List[j] = ordersListResponse{
				ProductID:    item.ProductID,
				ProductName:  item.ProductName,
				ProductPrice: item.ProductPrice,
				Count:        item.Count,
			}
		}

	}

	NewResponse(c, http.StatusOK, output)
}
