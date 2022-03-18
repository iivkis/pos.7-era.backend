package myservice

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"gorm.io/gorm"
)

type CashChangesOutputModel struct {
	ID         uint
	Date       int64   `json:"date"` //unixmilli
	Total      float64 `json:"total"`
	Reason     string  `json:"reason"`
	Comment    string  `json:"comment"`
	SessionID  uint    `json:"session_id"`
	EmployeeID uint    `json:"employee_id"`
	OutletID   uint    `json:"outletID"`
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

	claims := mustGetEmployeeClaims(c)

	if !s.repo.Sessions.Exists(&repository.SessionModel{Model: gorm.Model{ID: input.SessionID}, EmployeeID: claims.EmployeeID}) {
		NewResponse(c, http.StatusBadRequest, errRecordNotFound("session undefined"))
		return
	}

	model := repository.CashChangesModel{
		Date:       input.Date,
		Total:      input.Total,
		Reason:     input.Reason,
		Comment:    input.Comment,
		SessionID:  input.SessionID,
		EmployeeID: claims.EmployeeID,
		OutletID:   claims.OutletID,
		OrgID:      claims.OrganizationID,
	}

	if err := s.repo.CashChanges.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, DefaultOutputModel{ID: model.ID})
}

type CashChangesGetAllInput struct {
	Start uint64 `form:"start"` //in unixmilli
	End   uint64 `form:"end"`   //in unixmilli
}

type CashChangesGetAllOutput []CashChangesOutputModel

//@Summary Получить всю информацию о снятии\вкладе денежных средств (в точке)
//@param type query CashChangesGetAllInput false "Принимаемый объект"
//@Success 201 {object} CashChangesGetAllOutput "список изменений баланса кассы"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Router /cashChanges [get]
func (s *CashChangesService) GetAll(c *gin.Context) {
	var query CashChangesGetAllInput
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	stdQuery := mustGetStdQuery(c)
	claims := mustGetEmployeeClaims(c)

	items, err := s.repo.CashChanges.FindWithPeriod(query.Start, query.End, &repository.CashChangesModel{OrgID: claims.OrganizationID, OutletID: stdQuery.OutletID})
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output = make(CashChangesGetAllOutput, len(*items))
	for i, item := range *items {
		output[i] = CashChangesOutputModel{
			ID:         item.ID,
			Date:       item.Date,
			Total:      item.Total,
			Reason:     item.Reason,
			Comment:    item.Comment,
			SessionID:  item.SessionID,
			EmployeeID: item.EmployeeID,
			OutletID:   item.OutletID,
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
	claims := c.MustGet("claims").(*authjwt.EmployeeClaims)
	sess, err := s.repo.Sessions.GetLastOpenByEmployeeID(claims.EmployeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusOK, CashChangesGetAllForCurrentSessionOutput{})
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	items, err := s.repo.CashChanges.Find(&repository.CashChangesModel{SessionID: sess.ID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusOK, CashChangesGetAllForCurrentSessionOutput{})
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output = make(CashChangesGetAllForCurrentSessionOutput, len(*items))
	for i, item := range *items {
		output[i] = CashChangesOutputModel{
			ID:         item.ID,
			Date:       item.Date,
			Total:      item.Total,
			Reason:     item.Reason,
			Comment:    item.Comment,
			SessionID:  item.SessionID,
			EmployeeID: item.EmployeeID,
			OutletID:   item.OutletID,
		}
	}

	NewResponse(c, http.StatusOK, output)
}
