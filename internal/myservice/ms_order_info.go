package myservice

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"gorm.io/gorm"
)

type OrderInfoOutputModel struct {
	ID           uint   `json:"id"`
	PayType      int    `json:"pay_type"`
	Date         int64  `json:"date"`
	EmployeeName string `json:"employee_name"`
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
	PayType      int    `json:"pay_type" binding:"min=1,max=3"`
	EmployeeName string `json:"employee_name" binding:"required"`
	Date         int64  `json:"date" binding:"min=1"`
}

type OrdersInfoCreateOutput struct {
	ID uint `json:"id"`
}

//@Summary Добавить order info
//@param type body OrdersInfoCreateInput false "Принимаемый объект"
//@Accept json
//@Success 201 {object} OrdersInfoCreateOutput "возвращает id созданного order info"
//@Router /orderInfo [post]
func (s *OrdersInfoService) Create(c *gin.Context) {
	var input OrdersInfoCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	sess, err := s.repo.Sessions.GetByEmployeeID(c.MustGet("claims_employee_id").(uint))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound())
			return
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	newModel := repository.OrderInfoModel{
		PayType:      input.PayType,
		Date:         input.Date,
		EmployeeName: input.EmployeeName,
		SessionID:    sess.ID,
		OrgID:        c.MustGet("claims_org_id").(uint),
		OutletID:     c.MustGet("claims_outlet_id").(uint),
	}
	if err = s.repo.OrdersInfo.Create(&newModel); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := OrdersInfoCreateOutput{ID: newModel.ID}
	NewResponse(c, http.StatusCreated, output)
}

type OrdersInfoGetAllForOrgOutput []OrderInfoOutputModel

//@Summary Получить список order info организации
//@Accept json
//@Success 200 {object} OrdersInfoGetAllForOrgOutput "список order info"
//@Router /orderInfo [get]
func (s *OrdersInfoService) GetAllForOrg(c *gin.Context) {
	list, err := s.repo.OrdersInfo.FindAllForOrg(c.MustGet("claims_org_id").(uint))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := make(OrdersInfoGetAllForOrgOutput, len(list))
	for i, item := range list {
		output[i] = OrderInfoOutputModel{
			ID:           item.ID,
			PayType:      item.PayType,
			Date:         item.Date,
			EmployeeName: item.EmployeeName,
			SessionID:    item.SessionID,
			OutletID:     item.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}
