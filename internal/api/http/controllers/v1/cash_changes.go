package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type cashChangesResponseModel struct {
	ID         uint `json:"id" mapstructure:"id"`
	SessionID  uint `json:"session_id" mapstructure:"session_id"`
	EmployeeID uint `json:"employee_id" mapstructure:"employee_id"`
	OutletID   uint `json:"outletID" mapstructure:"outletID"`

	Date    int64   `json:"date" mapstructure:"date"` //unixmilli
	Total   float64 `json:"total" mapstructure:"total"`
	Reason  string  `json:"reason" mapstructure:"reason"`
	Comment string  `json:"comment" mapstructure:"comment"`
}

type cashChanges struct {
	repo *repository.Repository
}

func newCashChanges(repo *repository.Repository) *cashChanges {
	return &cashChanges{
		repo: repo,
	}
}

type cashChangesCreateBody struct {
	SessionID uint `json:"session_id"`

	Date    int64   `json:"date"` //unixmilli
	Total   float64 `json:"total"`
	Reason  string  `json:"reason" binding:"required"`
	Comment string  `json:"comment"`
}

// @Summary Добавить информацию о снятии\вкладе денежных средств
// @Param type body cashChangesCreateBody false "Принимаемый объект"
// @Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
// @Router /cashChanges [post]
func (s *cashChanges) Create(c *gin.Context) {
	var input cashChangesCreateBody
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
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, DefaultOutputModel{ID: model.ID})
}

type cashChangesGetAllQuery struct {
	Start uint64 `form:"start"` //in unixmilli
	End   uint64 `form:"end"`   //in unixmilli
}

type cashChangesGetAllResponse []cashChangesResponseModel

// @Summary Получить всю информацию о снятии\вкладе денежных средств (в точке)
// @Param type query cashChangesGetAllQuery false "query"
// @Success 200 {object} cashChangesGetAllResponse "список изменений баланса кассы"
// @Router /cashChanges [get]
func (s *cashChanges) GetAll(c *gin.Context) {
	var query cashChangesGetAllQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.CashChangesModel{
		OrgID:    claims.OrganizationID,
		OutletID: stdQuery.OutletID,
	}

	if claims.HasRole(repository.R_OWNER) {
		if stdQuery.OrgID != 0 && s.repo.Invitation.Exists(&repository.InvitationModel{OrgID: claims.OrganizationID, AffiliateOrgID: stdQuery.OrgID}) {
			where.OrgID = stdQuery.OrgID
		}
	}

	items, err := s.repo.CashChanges.FindWithPeriod(query.Start, query.End, where)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	var output = make(cashChangesGetAllResponse, len(*items))
	for i, item := range *items {
		output[i] = cashChangesResponseModel{
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

type cashChangesGetAllForCurrentSessionResponse []cashChangesResponseModel

// @Summary Получить информацию о снятии\вкладе денежных средств, которые были воспроизведены в текущей сессии (в точке)
// @Description берет последнюю открытую сессию (т.е. текущую сессию) сотрудника и по этой сессии ищет записи об изменении баланса кассы
// @Success 200 {object} cashChangesGetAllForCurrentSessionResponse "список изменений баланса кассы (по текущей сессии)"
// @Router /cashChanges.CurrentSession [get]
func (s *cashChanges) GetAllForCurrentSession(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)
	sess, err := s.repo.Sessions.GetLastOpenByEmployeeID(claims.EmployeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusOK, cashChangesGetAllForCurrentSessionResponse{})
		} else {
			NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		}
		return
	}

	items, err := s.repo.CashChanges.Find(&repository.CashChangesModel{SessionID: sess.ID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NewResponse(c, http.StatusOK, cashChangesGetAllForCurrentSessionResponse{})
		} else {
			NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		}
		return
	}

	var output = make(cashChangesGetAllForCurrentSessionResponse, len(*items))
	for i, item := range *items {
		output[i] = cashChangesResponseModel{
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
