package myservice

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"gorm.io/gorm"
)

type SessionService interface {
	OpenOrClose(c *gin.Context)
	GetAll(c *gin.Context)
	GetLastForOutlet(с *gin.Context)
}

type sessions struct {
	repo repository.Repository
}

type sessionOutputModel struct {
	ID         uint `json:"id"`
	EmployeeID uint `json:"employee_id"`
	OutletID   uint `json:"outlet_id"`

	CashOpen  float64 `json:"cash_open"`
	CashClose float64 `json:"cash_close"`

	DateOpen  string `json:"date_open"`
	DateClose string `json:"date_close"`
}

func newSessionService(repo repository.Repository) *sessions {
	return &sessions{
		repo: repo,
	}
}

type openOrCloseSessionInput struct {
	Action string  `json:"action" binding:"required"` // "open" or "close"
	Cash   float64 `json:"cash"`
}

//@Summary Открыть или закрыть сессию
//@Description Открывает сессию с id указанным в jwt токен
//@Description поле `action` принимает два параметра `open` (для открытия сессии) и `close` (для закрытия сессии)
//@param type body openOrCloseSessionInput false "Принимаемый объект"
//@Success 201 {object} object "возвращает пустой объект"
//@Router /sessions [post]
func (s *sessions) OpenOrClose(c *gin.Context) {
	var input openOrCloseSessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	switch input.Action {
	case "open":
		{
			sess := repository.SessionModel{
				CashSessionOpen: input.Cash,
				EmployeeID:      c.MustGet("claims_employee_id").(uint),
				OutletID:        c.MustGet("claims_outlet_id").(uint),
			}
			if err := s.repo.Sessions.Open(&sess); err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
				return
			}
			NewResponse(c, http.StatusOK, nil)
		}
	case "close":
		{
			if err := s.repo.Sessions.CloseByEmployeeID(c.MustGet("claims_employee_id").(uint), input.Cash); err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
				return
			}
			NewResponse(c, http.StatusOK, nil)
		}
	default:
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("action can be only `open` or `close` value"))
		return
	}
}

type getAllSessionsOutput []sessionOutputModel

//@Summary Список всех сессий
//@Description Метод позволяет получить список всех сессий
//@Produce json
//@Success 200 {object} getAllSessionsOutput "Возвращает массив сессий"
//@Failure 500 {object} serviceError
//@Router /sessions [get]
func (s *sessions) GetAll(c *gin.Context) {
	models, err := s.repo.Sessions.GetAllUnscoped()
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output getAllSessionsOutput = make(getAllSessionsOutput, len(models))
	for i, sess := range models {
		var dateClose string
		if !sess.DateClose.Time.IsZero() {
			dateClose = sess.DateClose.Time.String()
		}

		output[i] = sessionOutputModel{
			ID:         sess.ID,
			EmployeeID: sess.EmployeeID,
			OutletID:   sess.OutletID,
			CashOpen:   sess.CashSessionOpen,
			CashClose:  sess.CashSessionClose,
			DateOpen:   sess.DateOpen.String(),
			DateClose:  dateClose,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

//@Summary Последняя закрытая сессия торговой точки (к которой привязан jwt токен)
//@Description Метод позволяет получить последнюю сессию торговой точки, к которой привязан jwt токен
//@Produce json
//@Success 200 {object} sessionOutputModel "Возвращает последнюю закрытую сессию точки продаж"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /sessions/last [get]
func (s *sessions) GetLastForOutlet(c *gin.Context) {
	sess, err := s.repo.Sessions.GetLastForOutlet(c.MustGet("claims_outlet_id").(uint))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusBadRequest, errRecordNotFound())
			return
		}
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := sessionOutputModel{
		ID:         sess.ID,
		EmployeeID: sess.EmployeeID,
		OutletID:   sess.OutletID,
		CashOpen:   sess.CashSessionOpen,
		CashClose:  sess.CashSessionClose,
		DateOpen:   sess.DateOpen.String(),
		DateClose:  sess.DateClose.Time.String(),
	}
	NewResponse(c, http.StatusOK, output)
}
