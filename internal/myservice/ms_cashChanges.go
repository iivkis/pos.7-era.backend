package myservice

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type CashChangesOutputModel struct {
	ID        uint
	Date      int64   `json:"date"` //unixmilli
	Total     float64 `json:"total"`
	Reason    string  `json:"reason"`
	Comment   string  `json:"comment"`
	SessionID uint    `json:"session_id"`
	OutletID  uint    `json:"outletID"`
}

type CashChangesService struct {
	repo *repository.Repository
}

func newCashChangesService(repo *repository.Repository) *CashChangesService {
	return &CashChangesService{
		repo: repo,
	}
}

type CashChangesCreateInput struct {
	Date      int64   `json:"date"` //unixmilli
	Total     float64 `json:"total"`
	Reason    string  `json:"reason" binding:"required"`
	Comment   string  `json:"comment"`
	SessionID uint    `json:"session_id"`
}

//@Summary Добавить информацию о снятии\вкладе денежных средств
//@Description параметр `date` указывается в формате unixmilli
//@param type body CashChangesCreateInput false "Принимаемый объект"
//@Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Router /cashChanges [post]
func (s *CashChangesService) Create(c *gin.Context) {
	var input CashChangesCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	if !s.repo.Sessions.ExistsWithEmployeeID(input.SessionID, c.MustGet("claims_employee_id")) {
		NewResponse(c, http.StatusBadRequest, errRecordNotFound("session undefined"))
		return
	}

	m := repository.CashChangesModel{
		Date:       input.Date,
		Total:      input.Total,
		Reason:     input.Reason,
		Comment:    input.Comment,
		SessionID:  input.SessionID,
		EmployeeID: c.MustGet("claims_employee_id").(uint),
		OutletID:   c.MustGet("claims_outlet_id").(uint),
		OrgID:      c.MustGet("claims_org_id").(uint),
	}

	if err := s.repo.CashChanges.Create(&m); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, DefaultOutputModel{ID: m.ID})
}

type CashChangesGetAllForOutletInput struct {
	Start uint64 `json:"start"` //in unixmilli
	End   uint64 `json:"end"`   //in unixmilli
}

type CashChangesGetAllForOutletOutput []CashChangesOutputModel

//@Summary Получить всю информацию о снятии\вкладе денежных средств (в точке)
//@param type query CashChangesGetAllForOutletInput false "Принимаемый объект"
//@Success 201 {object} CashChangesGetAllForOutletOutput "список изменений баланса кассы"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Router /cashChanges [get]
func (s *CashChangesService) GetAllForOutlet(c *gin.Context) {
	var query CashChangesGetAllForOutletInput
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	items, err := s.repo.CashChanges.FindAllByOutletIDWithPeriod(query.Start, query.End, c.MustGet("claims_outlet_id"))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output = make(CashChangesGetAllForOutletOutput, len(items))
	for i, item := range items {
		output[i] = CashChangesOutputModel{
			ID:        item.ID,
			Date:      item.Date,
			Total:     item.Total,
			Reason:    item.Reason,
			Comment:   item.Comment,
			SessionID: item.SessionID,
			OutletID:  item.OutletID,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

type CashChangesGetAllForCurrentSessionOutput []CashChangesOutputModel

//@Summary Получить информацию о снятии\вкладе денежных средств, которые были воспроизведены в текущей сессии (в точке)
//@Description берет последнюю открытую сессию (т.е. текущую сессию) сотрудника и по этой сессии ищет записи об изменении баланса кассы
//@Success 201 {object} CashChangesGetAllForCurrentSessionOutput "список изменений баланса кассы (по текущей сессии)"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Router /cashChanges.CurrentSession [get]
func (s *CashChangesService) GetAllForCurrentSession(c *gin.Context) {
	sess, err := s.repo.Sessions.GetLastOpenByEmployeeID(c.MustGet("claims_employee_id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusOK, make(CashChangesGetAllForCurrentSessionOutput, 0))
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	items, err := s.repo.CashChanges.FindAllBySessionID(sess.ID)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output = make(CashChangesGetAllForCurrentSessionOutput, len(items))
	for i, item := range items {
		output[i] = CashChangesOutputModel{
			ID:        item.ID,
			Date:      item.Date,
			Total:     item.Total,
			Reason:    item.Reason,
			Comment:   item.Comment,
			SessionID: item.SessionID,
			OutletID:  item.OutletID,
		}
	}

	NewResponse(c, http.StatusOK, output)
}
