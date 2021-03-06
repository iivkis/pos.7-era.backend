package myservice

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type SessionsService struct {
	repo *repository.Repository
}

type SessionOutputModel struct {
	ID uint `json:"id"`

	СashEarned float64 `json:"cash_earned"`
	BankEarned float64 `json:"bank_earned"`

	NumberOfReceipts int `json:"number_of_receipts"`

	CashOpen  float64 `json:"cash_open"`
	CashClose float64 `json:"cash_close"`

	DateOpen  int64 `json:"date_open"`  //unixmilli
	DateClose int64 `json:"date_close"` //unixmilli

	EmployeeID uint `json:"employee_id"`
	OutletID   uint `json:"outlet_id"`
}

func newSessionsService(repo *repository.Repository) *SessionsService {
	return &SessionsService{
		repo: repo,
	}
}

type SessionsOpenOrCloseInput struct {
	Action string `json:"action" binding:"required"` // "open" or "close"

	Date int64   `json:"date" binding:"min=1"`
	Cash float64 `json:"cash"`

	CashEarned float64 `json:"cash_earned"`
	BankEarned float64 `json:"bank_earned"`
}

type SessionOpenOrCloseOutput struct {
	ID         uint `json:"id"`
	EmployeeID uint `json:"employee_id"`
}

//@Summary Открыть или закрыть сессию в точке
//@Description Открывает сессию с id указанным в jwt токен.
//@Description - Поле `action` принимает два параметра `open` (для открытия сессии) и `close` (для закрытия сессии)
//@param type body SessionsOpenOrCloseInput false "Принимаемый объект"
//@Success 201 {object} SessionOpenOrCloseOutput "возвращает id созданной записи"
//@Router /sessions [post]
func (s *SessionsService) OpenOrClose(c *gin.Context) {
	var input SessionsOpenOrCloseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	switch input.Action {
	case "open":
		{
			sess := repository.SessionModel{
				CashSessionOpen: input.Cash,
				DateOpen:        input.Date,
				EmployeeID:      claims.EmployeeID,
				OutletID:        claims.OutletID,
				OrgID:           claims.OrganizationID,
			}
			if err := s.repo.Sessions.Open(&sess); err != nil {
				if errors.Is(err, repository.ErrSessionAlreadyOpen) {
					NewResponse(c, http.StatusBadRequest, errRecordAlreadyExists(err.Error()))
					return
				}
				NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
				return
			}

			if err := s.repo.Employees.SetOnline(claims.EmployeeID); err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
				return
			}

			NewResponse(c, http.StatusOK, SessionOpenOrCloseOutput{ID: sess.ID, EmployeeID: sess.EmployeeID})
		}

	case "close":
		{
			lastOpenEmployeeSession, err := s.repo.Sessions.GetLastOpenByEmployeeID(claims.EmployeeID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					NewResponse(c, http.StatusBadRequest, errRecordNotFound("undefined open session"))
					return
				}
				NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
				return
			}

			NumberOfReceipts, err := s.repo.OrdersInfo.Count(&repository.OrderInfoModel{SessionID: lastOpenEmployeeSession.ID})
			if err != nil {
				NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
				return
			}

			sess := repository.SessionModel{
				DateClose:        input.Date,
				CashSessionClose: input.Cash,
				BankEarned:       input.BankEarned,
				CashEarned:       input.CashEarned,
				NumberOfReceipts: int(NumberOfReceipts),
			}

			if err := s.repo.Sessions.Close(claims.EmployeeID, &sess); err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
				return
			}

			if err := s.repo.Employees.SetOffline(claims.EmployeeID); err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
				return
			}
			NewResponse(c, http.StatusOK, SessionOpenOrCloseOutput{ID: sess.ID, EmployeeID: sess.EmployeeID})
		}
	default:
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("action can be only `open` or `close` value"))
		return
	}
}

type SessionsGetAllInput struct {
	Start uint64 `form:"start"` //in unixmilli
	End   uint64 `form:"end"`   //in unixmilli
}

type SessionsGetAllOutput []SessionOutputModel

//@Summary Список всех сессий точки
//@Description Метод позволяет получить список всех сессий точки
//@Param type query SessionsGetAllInput false "принимаемые поля"
//@Success 200 {object} SessionsGetAllOutput "Возвращает массив сессий точки"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /sessions [get]
func (s *SessionsService) GetAll(c *gin.Context) {
	var query SessionsGetAllInput
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}
	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.SessionModel{
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER) {
		if stdQuery.OrgID != 0 && s.repo.Invitation.Exists(&repository.InvitationModel{OrgID: claims.OrganizationID, AffiliateOrgID: stdQuery.OrgID}) {
			where.OrgID = stdQuery.OrgID
		}
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	sessions, err := s.repo.Sessions.FindWithPeriod(query.Start, query.End, where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	var output SessionsGetAllOutput = make(SessionsGetAllOutput, len(*sessions))
	for i, sess := range *sessions {
		output[i] = SessionOutputModel{
			ID:         sess.ID,
			EmployeeID: sess.EmployeeID,
			OutletID:   sess.OutletID,

			CashOpen:  sess.CashSessionOpen,
			CashClose: sess.CashSessionClose,

			СashEarned:       sess.CashEarned,
			BankEarned:       sess.BankEarned,
			NumberOfReceipts: sess.NumberOfReceipts,

			DateOpen:  sess.DateOpen,
			DateClose: sess.DateClose,
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
	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	outletID := claims.OutletID
	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			outletID = stdQuery.OutletID
		}
	}

	sess, err := s.repo.Sessions.GetLastClosedForOutlet(outletID)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := SessionOutputModel{
		ID:         sess.ID,
		EmployeeID: sess.EmployeeID,
		OutletID:   sess.OutletID,

		CashOpen:  sess.CashSessionOpen,
		CashClose: sess.CashSessionClose,

		СashEarned:       sess.CashEarned,
		BankEarned:       sess.BankEarned,
		NumberOfReceipts: sess.NumberOfReceipts,

		DateOpen:  sess.DateOpen,
		DateClose: sess.DateClose,
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
	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	outletID := claims.OutletID
	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			outletID = stdQuery.OutletID
		}
	}

	sess, err := s.repo.Sessions.GetLastForOutlet(outletID)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := SessionOutputModel{
		ID:         sess.ID,
		EmployeeID: sess.EmployeeID,
		OutletID:   sess.OutletID,

		CashOpen:  sess.CashSessionOpen,
		CashClose: sess.CashSessionClose,

		СashEarned:       sess.CashEarned,
		BankEarned:       sess.BankEarned,
		NumberOfReceipts: sess.NumberOfReceipts,

		DateOpen:  sess.DateOpen,
		DateClose: sess.DateClose,
	}

	NewResponse(c, http.StatusOK, output)
}

//@Summary Последняя сессия текущего юзера (к которой привязан jwt токен)
//@Success 200 {object} SessionOutputModel "Возвращает последнюю сессию текущего юзера"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /sessions.Last.Me [get]
func (s *SessionsService) GetLastForMe(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	sess, err := s.repo.Sessions.GetLastForEmployeeByID(claims.EmployeeID)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := SessionOutputModel{
		ID:         sess.ID,
		EmployeeID: sess.EmployeeID,
		OutletID:   sess.OutletID,

		CashOpen:  sess.CashSessionOpen,
		CashClose: sess.CashSessionClose,

		СashEarned:       sess.CashEarned,
		BankEarned:       sess.BankEarned,
		NumberOfReceipts: sess.NumberOfReceipts,

		DateOpen:  sess.DateOpen,
		DateClose: sess.DateClose,
	}

	NewResponse(c, http.StatusOK, output)
}
