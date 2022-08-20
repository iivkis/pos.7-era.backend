package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type orderInfoResponseModel struct {
	ID        uint `json:"id" mapstructure:"id"`
	SessionID uint `json:"session_id" mapstructure:"session_id"`
	OutletID  uint `json:"outlet_id" mapstructure:"outlet_id"`

	Date         int64  `json:"date" mapstructure:"date"`
	PayType      int    `json:"pay_type" mapstructure:"pay_type"`
	EmployeeName string `json:"employee_name" mapstructure:"employee_name"`

	IsDelete bool `json:"is_delete" mapstructure:"is_delete"`
}

type orderInfo struct {
	repo *repository.Repository
}

func newOrderInfo(repo *repository.Repository) *orderInfo {
	return &orderInfo{
		repo: repo,
	}
}

type orderInfoCreateBody struct {
	SessionID uint `json:"session_id" binding:"min=1"`

	Date         int64  `json:"date" binding:"min=1"`
	PayType      int    `json:"pay_type" binding:"min=0,max=1"`
	EmployeeName string `json:"employee_name" binding:"required"`
}

// @Summary Добавить orderInfo (список завершенных заказов)
// @param type body orderInfoCreateBody false "Принимаемый объект"
// @Success 201 {object} DefaultOutputModel "возвращает id созданного order info"
// @Router /orderInfo [post]
func (s *orderInfo) Create(c *gin.Context) {
	var body orderInfoCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	//check session
	sess, err := s.repo.Sessions.FindFirts(&repository.SessionModel{Model: gorm.Model{ID: body.SessionID}, OutletID: claims.OutletID})
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

	//create model

	model := repository.OrderInfoModel{
		PayType:      body.PayType,
		Date:         body.Date,
		EmployeeName: body.EmployeeName,
		SessionID:    body.SessionID,
		OrgID:        claims.OrganizationID,
		OutletID:     claims.OutletID,
	}

	if err = s.repo.OrdersInfo.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type orderInfoGetAllQuery struct {
	SessionID uint `form:"session_id"`
}

type orderInfoGetAllResponse []orderInfoResponseModel

// @Summary Получить список завершенных заказов (orderInfo)
// @Param type query orderInfoGetAllQuery false "query"
// @Success 200 {object} orderInfoGetAllResponse "список завершенных заказов"
// @Router /orderInfo [get]
func (s *orderInfo) GetAll(c *gin.Context) {
	var query orderInfoGetAllQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStandartQuery(c)

	where := &repository.OrderInfoModel{
		OrgID:     claims.OrganizationID,
		OutletID:  claims.OutletID,
		SessionID: query.SessionID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	list, err := s.repo.OrdersInfo.FindUnscoped(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := make(orderInfoGetAllResponse, len(*list))
	for i, item := range *list {
		output[i] = orderInfoResponseModel{
			ID:           item.ID,
			PayType:      item.PayType,
			Date:         item.Date,
			EmployeeName: item.EmployeeName,
			IsDelete:     !item.DeletedAt.Time.IsZero(),
			SessionID:    item.SessionID,
			OutletID:     item.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}

// @Summary Удалить orderInfo в точке по его id
// @Success 200 {object} object "возвращает пустой объект"
// @Router /orderInfo/:id [delete]
func (s *orderInfo) Delete(c *gin.Context) {
	orderInfoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStandartQuery(c)

	where := &repository.OrderInfoModel{
		Model:    gorm.Model{ID: uint(orderInfoID)},
		OutletID: claims.OutletID,
		OrgID:    claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	//есть ли orderInfo в точке и организации
	if !s.repo.OrdersInfo.Exists(where) {
		NewResponse(c, http.StatusBadRequest, errRecordNotFound("undefined `order_info` with this `id`"))
		return
	}

	orderLists, err := s.repo.OrdersList.Find(&repository.OrderListModel{OrderInfoID: where.ID})
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	for _, orderList := range *orderLists {
		if err := s.repo.ProductsWithIngredients.AdditionIngredients(orderList.ProductID, orderList.Count); err != nil {
			NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
			return
		}
	}

	if err := s.repo.OrdersList.Delete(&repository.OrderListModel{OrderInfoID: uint(orderInfoID)}); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	if err := s.repo.OrdersInfo.Delete(where); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

// @Summary Восстановить orderInfo в точке по его id
// @Success 200 {object} object "object"
// @Router /orderInfo/:id [post]
func (s *orderInfo) Recovery(c *gin.Context) {
	orderInfoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStandartQuery(c)

	where := &repository.OrderInfoModel{
		Model:    gorm.Model{ID: uint(orderInfoID)},
		OutletID: claims.OutletID,
		OrgID:    claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER) {
		if stdQuery.OrgID != 0 && s.repo.Invitation.Exists(&repository.InvitationModel{OrgID: claims.OrganizationID, AffiliateOrgID: stdQuery.OrgID}) {
			where.OrgID = stdQuery.OrgID
		}
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	//check orderInfo
	{
		orderInfo, err := s.repo.OrdersInfo.FindFirstUnscoped(where)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				NewResponse(c, http.StatusBadRequest, errRecordNotFound("undefined `order_info` with this `id`"))
			} else {
				NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
			}
			return
		}

		if orderInfo.DeletedAt.Time.IsZero() {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound("record already recovered"))
			return
		}
	}

	orderLists, err := s.repo.OrdersList.FindUnscoped(&repository.OrderListModel{OrderInfoID: where.ID, OutletID: where.OutletID, OrgID: where.OrgID})
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	for _, orderList := range *orderLists {
		if err := s.repo.ProductsWithIngredients.SubractionIngredients(orderList.ProductID, orderList.Count); err != nil {
			NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
			return
		}
	}

	if err := s.repo.OrdersList.Recovery(&repository.OrderListModel{OrderInfoID: uint(orderInfoID)}); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	if err := s.repo.OrdersInfo.Recovery(where); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}
