package myservice

import (
	"errors"
	"net/http"

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
	ID         uint `json:"id"`
	EmployeeID uint `json:"employee_id"`
}

//@Summary Открыть или закрыть сессию в точке
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
				DateOpen:        input.Date,
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
			NewResponse(c, http.StatusOK, SessionOpenOrCloseOutput{ID: sess.ID, EmployeeID: sess.EmployeeID})
		}
	case "close":
		{
			sess := repository.SessionModel{
				DateClose:        input.Date,
				CashSessionClose: input.Cash,
			}
			err := s.repo.Sessions.Close(c.MustGet("claims_employee_id"), &sess)
			if err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
				return
			}
			NewResponse(c, http.StatusOK, SessionOpenOrCloseOutput{ID: sess.ID, EmployeeID: sess.EmployeeID})
		}
	default:
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("action can be only `open` or `close` value"))
		return
	}
}

type SessionsGetAllForOutletOutput []SessionOutputModel

//@Summary Список всех сессий точки
//@Description Метод позволяет получить список всех сессий точки
//@Success 200 {object} SessionsGetAllForOutletOutput "Возвращает массив сессий точки"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /sessions [get]
func (s *SessionsService) GetAllForOutlet(c *gin.Context) {
	sessions, err := s.repo.Sessions.GetAllByOutletID(c.MustGet("claims_outlet_id").(uint))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output SessionsGetAllForOutletOutput = make(SessionsGetAllForOutletOutput, len(sessions))
	for i, sess := range sessions {
		output[i] = SessionOutputModel{
			ID:         sess.ID,
			EmployeeID: sess.EmployeeID,
			OutletID:   sess.OutletID,
			CashOpen:   sess.CashSessionOpen,
			CashClose:  sess.CashSessionClose,
			DateOpen:   sess.DateOpen,
			DateClose:  sess.DateClose,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

//@Summary Последняя закрытая сессия торговой точки (к которой привязан jwt токен)
//@Description Метод позволяет получить последнюю сессию торговой точки, к которой привязан jwt токен
//@Accept json
//@Produce json
//@Success 200 {object} SessionOutputModel "Возвращает последнюю закрытую сессию точки продаж"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /sessions.Last.Closed [get]
func (s *SessionsService) GetLastClosedForOutlet(c *gin.Context) {
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
		DateOpen:   sess.DateOpen,
		DateClose:  sess.DateClose,
	}

	NewResponse(c, http.StatusOK, output)
}

//@Summary Последняя сессия торговой точки (к которой привязан jwt токен)
//@Description Метод позволяет получить последнюю сессию торговой точки (не важно, открытая или закрытая), к которой привязан jwt токен
//@Success 200 {object} SessionOutputModel "Возвращает последнюю закрытую сессию точки продаж"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /sessions.Last [get]
func (s *SessionsService) GetLastForOutlet(c *gin.Context) {
	sess, err := s.repo.Sessions.GetLastForOutlet(c.MustGet("claims_outlet_id").(uint))
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
		DateOpen:   sess.DateOpen,
		DateClose:  sess.DateClose,
	}

	NewResponse(c, http.StatusOK, output)
}
