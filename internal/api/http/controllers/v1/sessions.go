package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type sessions struct {
	repo *repository.Repository
}

type sessionsResponseModel struct {
	ID         uint `json:"id" mapstructure:"id"`
	EmployeeID uint `json:"employee_id" mapstructure:"employee_id"`
	OutletID   uint `json:"outlet_id" mapstructure:"outlet_id"`

	CashEarned float64 `json:"cash_earned" mapstructure:"cash_earned"` //заработок за наличную оплату
	BankEarned float64 `json:"bank_earned" mapstructure:"bank_earned"` //заработок по оплате через карту

	CashOpen  float64 `json:"cash_open" mapstructure:"cash_open"`   //баланс в момент открытия кассы
	CashClose float64 `json:"cash_close" mapstructure:"cash_close"` //баланс в момент закрытия кассы

	DateOpen  int64 `json:"date_open" mapstructure:"date_open"`   //UnixMilli
	DateClose int64 `json:"date_close" mapstructure:"date_close"` //UnixMilli

	NumberOfReceipts int `json:"number_of_receipts" mapstructure:"number_of_receipts"` //количество чеков
}

func newSessions(repo *repository.Repository) *sessions {
	return &sessions{
		repo: repo,
	}
}

type sessionsOpenOrCloseBody struct {
	Action string `json:"action" binding:"required"` // "open" or "close"

	Date int64   `json:"date" binding:"min=1"`
	Cash float64 `json:"cash"`

	CashEarned float64 `json:"cash_earned"`
	BankEarned float64 `json:"bank_earned"`
}

type sessionsOpenOrCloseResponse struct {
	ID         uint `json:"id" mapstructure:"id"`
	EmployeeID uint `json:"employee_id" mapstructure:"employee_id"`
}

// @Summary Открыть или закрыть сессию в точке
// @Description Поле `action` принимает два параметра: `open` (для открытия сессии) и `close` (для закрытия сессии)
// @Param type body sessionsOpenOrCloseBody false "object"
// @Success 200 {object} sessionsOpenOrCloseResponse "object"
// @Router /sessions [post]
func (s *sessions) OpenOrClose(c *gin.Context) {
	var body sessionsOpenOrCloseBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	switch body.Action {
	case "open":
		{
			sess := repository.SessionModel{
				CashSessionOpen: body.Cash,
				DateOpen:        body.Date,
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

			NewResponse(c, http.StatusOK, sessionsOpenOrCloseResponse{ID: sess.ID, EmployeeID: sess.EmployeeID})
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

			numberOfReceipts, err := s.repo.OrdersInfo.Count(&repository.OrderInfoModel{SessionID: lastOpenEmployeeSession.ID})
			if err != nil {
				NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
				return
			}

			sess := repository.SessionModel{
				DateClose:        body.Date,
				CashSessionClose: body.Cash,
				BankEarned:       body.BankEarned,
				CashEarned:       body.CashEarned,
				NumberOfReceipts: int(numberOfReceipts),
			}

			if err := s.repo.Sessions.Close(claims.EmployeeID, &sess); err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
				return
			}

			if err := s.repo.Employees.SetOffline(claims.EmployeeID); err != nil {
				NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
				return
			}
			NewResponse(c, http.StatusOK, sessionsOpenOrCloseResponse{ID: sess.ID, EmployeeID: sess.EmployeeID})
		}
	default:
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("action can be only `open` or `close` value"))
		return
	}
}

type sessionsGetAllQuery struct {
	Start uint64 `form:"start"` //in unixmilli
	End   uint64 `form:"end"`   //in unixmilli
}

type sessionsGetAllResponse []sessionsResponseModel

// @Summary Список всех сессий точки
// @Description Метод позволяет получить список всех сессий точки
// @Param type query sessionsGetAllQuery false "принимаемые поля"
// @Success 200 {object} sessionsGetAllResponse "Возвращает массив сессий точки"
// @Router /sessions [get]
func (s *sessions) GetAll(c *gin.Context) {
	var query sessionsGetAllQuery
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

	var output sessionsGetAllResponse = make(sessionsGetAllResponse, len(*sessions))
	for i, sess := range *sessions {
		output[i] = sessionsResponseModel{
			ID:         sess.ID,
			EmployeeID: sess.EmployeeID,
			OutletID:   sess.OutletID,

			CashOpen:  sess.CashSessionOpen,
			CashClose: sess.CashSessionClose,

			CashEarned:       sess.CashEarned,
			BankEarned:       sess.BankEarned,
			NumberOfReceipts: sess.NumberOfReceipts,

			DateOpen:  sess.DateOpen,
			DateClose: sess.DateClose,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

// @Summary Последняя закрытая сессия торговой точки (к которой привязан jwt токен)
// @Description Метод позволяет получить последнюю сессию торговой точки, к которой привязан jwt токен
// @Success 200 {object} sessionsResponseModel "Возвращает последнюю закрытую сессию точки продаж"
// @Router /sessions.Last.Closed [get]
func (s *sessions) GetLastClosedForOutlet(c *gin.Context) {
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

	output := sessionsResponseModel{
		ID:         sess.ID,
		EmployeeID: sess.EmployeeID,
		OutletID:   sess.OutletID,

		CashOpen:  sess.CashSessionOpen,
		CashClose: sess.CashSessionClose,

		CashEarned:       sess.CashEarned,
		BankEarned:       sess.BankEarned,
		NumberOfReceipts: sess.NumberOfReceipts,

		DateOpen:  sess.DateOpen,
		DateClose: sess.DateClose,
	}

	NewResponse(c, http.StatusOK, output)
}

// @Summary Последняя сессия торговой точки (к которой привязан jwt токен)
// @Description Метод позволяет получить последнюю сессию торговой точки (не важно, открытая или закрытая), к которой привязан jwt токен
// @Success 200 {object} sessionsResponseModel "Возвращает последнюю закрытую сессию точки продаж"
// @Router /sessions.Last [get]
func (s *sessions) GetLastForOutlet(c *gin.Context) {
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

	output := sessionsResponseModel{
		ID:         sess.ID,
		EmployeeID: sess.EmployeeID,
		OutletID:   sess.OutletID,

		CashOpen:  sess.CashSessionOpen,
		CashClose: sess.CashSessionClose,

		CashEarned:       sess.CashEarned,
		BankEarned:       sess.BankEarned,
		NumberOfReceipts: sess.NumberOfReceipts,

		DateOpen:  sess.DateOpen,
		DateClose: sess.DateClose,
	}

	NewResponse(c, http.StatusOK, output)
}

// @Summary Последняя сессия текущего юзера (к которой привязан jwt токен)
// @Success 200 {object} sessionsResponseModel "Возвращает последнюю сессию текущего юзера"
// @Router /sessions.Last.Me [get]
func (s *sessions) GetLastForMe(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	sess, err := s.repo.Sessions.GetLastForEmployeeByID(claims.EmployeeID)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := sessionsResponseModel{
		ID:         sess.ID,
		EmployeeID: sess.EmployeeID,
		OutletID:   sess.OutletID,

		CashOpen:  sess.CashSessionOpen,
		CashClose: sess.CashSessionClose,

		CashEarned:       sess.CashEarned,
		BankEarned:       sess.BankEarned,
		NumberOfReceipts: sess.NumberOfReceipts,

		DateOpen:  sess.DateOpen,
		DateClose: sess.DateClose,
	}

	NewResponse(c, http.StatusOK, output)
}
