package myservice

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type OrderInfoOutputModel struct {
	ID           uint   `json:"id"`
	PayType      int    `json:"pay_type"`
	Date         int64  `json:"date"`
	EmployeeName string `json:"employee_name"`
	IsDelete     bool   `json:"is_delete"`
	SessionID    uint   `json:"session_id"`
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
	PayType      int    `json:"pay_type" binding:"min=0,max=2"`
	EmployeeName string `json:"employee_name" binding:"required"`
	Date         int64  `json:"date" binding:"min=1"`
}

//@Summary Добавить orderInfo (список завершенных заказов)
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

	claims := mustGetEmployeeClaims(c)

	sess, err := s.repo.Sessions.GetLastOpenByEmployeeID(claims.EmployeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound("you should open new session"))
			return
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	model := repository.OrderInfoModel{
		PayType:      input.PayType,
		Date:         input.Date,
		EmployeeName: input.EmployeeName,
		SessionID:    sess.ID,
		OrgID:        claims.OrganizationID,
		OutletID:     claims.OutletID,
	}
	if err = s.repo.OrdersInfo.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type OrdersInfoGetAllOutput []OrderInfoOutputModel

//@Summary Получить список завершенных заказов (orderInfo)
//@Accept json
//@Produce json
//@Success 200 {object} OrdersInfoGetAllOutput "список завершенных заказов"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /orderInfo [get]
func (s *OrdersInfoService) GetAllForOutlet(c *gin.Context) {
	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.OrderInfoModel{
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	list, err := s.repo.OrdersInfo.Find(where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := make(OrdersInfoGetAllOutput, len(*list))
	for i, item := range *list {
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
	orderInfoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.OrderInfoModel{
		Model: gorm.Model{ID: uint(orderInfoID)},
		OrgID: claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	if err := s.repo.OrdersInfo.Delete(where); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}
