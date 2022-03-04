package myservice

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type OrderInfoOutputModel struct {
	ID           uint   `json:"id"`
	PayType      int    `json:"pay_type"`
	Date         int64  `json:"date"`
	EmployeeName string `json:"employee_name"`
	SessionID    uint   `json:"session_id"`
	IsDelete     bool   `json:"is_delete"`
	OutletID     uint   `json:"outlet_id"`
}

type OrdersInfoService struct {
	repo *repository.Repository
}

func newOrdersInfoService(repo *repository.Repository) *OrdersInfoService {
	return &OrdersInfoService{
		repo: repo,
	}
}

type OrdersInfoCreateInput struct {
	PayType      int    `json:"pay_type" binding:"min=1,max=3"`
	EmployeeName string `json:"employee_name" binding:"required"`
	Date         int64  `json:"date" binding:"min=1"`
}

//@Summary Добавить orderInfo
//@param type body OrdersInfoCreateInput false "Принимаемый объект"
//@Accept json
//@Success 201 {object} DefaultOutputModel "возвращает id созданного order info"
//@Router /orderInfo [post]
func (s *OrdersInfoService) Create(c *gin.Context) {
	var input OrdersInfoCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	sess, err := s.repo.Sessions.GetLastOpenByEmployeeID(c.MustGet("claims_employee_id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound("you should open new session"))
			return
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	m := repository.OrderInfoModel{
		PayType:      input.PayType,
		Date:         input.Date,
		EmployeeName: input.EmployeeName,
		SessionID:    sess.ID,
		OrgID:        c.MustGet("claims_org_id").(uint),
		OutletID:     c.MustGet("claims_outlet_id").(uint),
	}
	if err = s.repo.OrdersInfo.Create(&m); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: m.ID})
}

type OrdersInfoGetAllForOutletOutput []OrderInfoOutputModel

//@Summary Получить список завершенных заказов точки (orderInfo)
//@Accept json
//@Produce json
//@Success 200 {object} OrdersInfoGetAllForOutletOutput "список завершенных заказов точки"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /orderInfo [get]
func (s *OrdersInfoService) GetAllForOutlet(c *gin.Context) {
	list, err := s.repo.OrdersInfo.FindAllByOutletID(c.MustGet("claims_outlet_id"))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := make(OrdersInfoGetAllForOutletOutput, len(list))
	for i, item := range list {
		output[i] = OrderInfoOutputModel{
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

//@Summary Удалить orderInfo в точке по его id
//@Success 200 {object} object "возвращает пустой объект"
//@Produce json
//@Accept json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /orderInfo/:id [delete]
func (s *OrdersInfoService) Delete(c *gin.Context) {
	err := s.repo.OrdersInfo.Delete(c.Param("id"), c.MustGet("claims_outlet_id"))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}
