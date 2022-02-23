package myservice

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

type SessionsService struct {
	repo *repository.Repository
}

type SessionOutputModel struct {
	ID         uint `json:"id"`
	EmployeeID uint `json:"employee_id"`
	OutletID   uint `json:"outlet_id"`

	CashOpen  float64 `json:"cash_open"`
	CashClose float64 `json:"cash_close"`

	DateOpen  int64 `json:"date_open"`
	DateClose int64 `json:"date_close"`
}

func newSessionsService(repo *repository.Repository) *SessionsService {
	return &SessionsService{
		repo: repo,
	}
}

type SessionsOpenOrCloseInput struct {
	Action string  `json:"action" binding:"required"` // "open" or "close"
	Date   int64   `json:"date"`
	Cash   float64 `json:"cash"`
}

type SessionOpenOrCloseOutput struct {
	ID uint `json:"id"`
}

//@Summary Открыть или закрыть сессию
//@Description Открывает сессию с id указанным в jwt токен.
//@Description - Поле `action` принимает два параметра `open` (для открытия сессии) и `close` (для закрытия сессии)
//@param type body SessionsOpenOrCloseInput false "Принимаемый объект"
//@Success 201 {object} object "возвращает пустой объект"
//@Router /sessions [post]
func (s *SessionsService) OpenOrClose(c *gin.Context) {
	var input SessionsOpenOrCloseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	switch input.Action {
	case "open":
		{
			sess := repository.SessionModel{
				CashSessionOpen: input.Cash,
				DateOpen:        time.UnixMilli(input.Date),
				EmployeeID:      c.MustGet("claims_employee_id").(uint),
				OutletID:        c.MustGet("claims_outlet_id").(uint),
				OrgID:           c.MustGet("claims_org_id").(uint),
			}
			if err := s.repo.Sessions.Open(&sess); err != nil {
				if errors.Is(err, repository.ErrSessionAlreadyOpen) {
					NewResponse(c, http.StatusBadRequest, errRecordAlreadyExists(err.Error()))
					return
				}
				NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
				return
			}
			NewResponse(c, http.StatusOK, SessionOpenOrCloseOutput{ID: sess.ID})
		}
	case "close":
		{
			id, err := s.repo.Sessions.CloseByEmployeeID(c.MustGet("claims_employee_id").(uint), time.UnixMilli(input.Date), input.Cash)
			if err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
				return
			}
			NewResponse(c, http.StatusOK, SessionOpenOrCloseOutput{ID: id})
		}
	default:
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("action can be only `open` or `close` value"))
		return
	}
}

type getAllSessionsOutput []SessionOutputModel

//@Summary Список всех сессий
//@Description Метод позволяет получить список всех сессий
//@Produce json
//@Success 200 {object} getAllSessionsOutput "Возвращает массив сессий"
//@Failure 500 {object} serviceError
//@Router /sessions [get]
func (s *SessionsService) GetAll(c *gin.Context) {
	sessions, err := s.repo.Sessions.GetAllUnscopedByOrgID(c.MustGet("claims_org_id").(uint))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output getAllSessionsOutput = make(getAllSessionsOutput, len(sessions))
	for i, sess := range sessions {
		var dateClose int64
		if !sess.DateClose.Time.IsZero() {
			dateClose = sess.DateClose.Time.UnixMilli()
		}

		output[i] = SessionOutputModel{
			ID:         sess.ID,
			EmployeeID: sess.EmployeeID,
			OutletID:   sess.OutletID,
			CashOpen:   sess.CashSessionOpen,
			CashClose:  sess.CashSessionClose,
			DateOpen:   sess.DateOpen.UnixMilli(),
			DateClose:  dateClose,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

//@Summary Последняя закрытая сессия торговой точки (к которой привязан jwt токен)
//@Description Метод позволяет получить последнюю сессию торговой точки, к которой привязан jwt токен
//@Produce json
//@Success 200 {object} SessionOutputModel "Возвращает последнюю закрытую сессию точки продаж"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /sessions/last [get]
func (s *SessionsService) GetLastForOutlet(c *gin.Context) {
	sess, err := s.repo.Sessions.GetLastClosedForOutlet(c.MustGet("claims_outlet_id").(uint))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := SessionOutputModel{
		ID:         sess.ID,
		EmployeeID: sess.EmployeeID,
		OutletID:   sess.OutletID,
		CashOpen:   sess.CashSessionOpen,
		CashClose:  sess.CashSessionClose,
		DateOpen:   sess.DateOpen.UnixMilli(),
		DateClose:  sess.DateClose.Time.UnixMilli(),
	}

	NewResponse(c, http.StatusOK, output)
}
