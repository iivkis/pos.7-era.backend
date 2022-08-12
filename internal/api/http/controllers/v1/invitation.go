package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type invitationResponseModel struct {
	ID             uint `json:"id" mapstructure:"id"`                             //id инвайта
	AffiliateOrgID uint `json:"affiliate_org_id" mapstructure:"affiliate_org_id"` // приглашенная организация

	Code      string `json:"code" mapstructure:"code"`             // рандомный код (равен "", если активирован)
	ExpiresIn int64  `json:"expires_in" mapstructure:"expires_in"` // когда истекает инвайт (равен 0, если активирован)
}

type invitation struct {
	repo *repository.Repository
}

func newInvitation(repo *repository.Repository) *invitation {
	return &invitation{
		repo: repo,
	}
}

type invitationCreateResponse struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
}

// @Summary Создать код приглашения
// @Success 201 {object} invitationCreateResponse "возвращает id созданной записи и код приглашения"
// @Router /invites [post]
func (s *invitation) Create(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	//запретить, если оргнизация сама является дочерней
	if s.repo.Invitation.Exists(&repository.InvitationModel{AffiliateOrgID: claims.OrganizationID}) {
		NewResponse(c, http.StatusBadRequest, errPermissionDenided("an affiliated organization cannot create invitation codes"))
		return
	}

	if n, err := s.repo.Invitation.CountNotActivated(claims.OrganizationID); err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	} else if n >= 10 {
		NewResponse(c, http.StatusBadRequest, errRecordAlreadyExists("you can create up to 10 inactive invites"))
		return
	}

	invite := &repository.InvitationModel{
		OrgID: claims.OrganizationID,
	}

	if err := s.repo.Invitation.Create(invite); err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}

	output := invitationCreateResponse{
		ID:   invite.ID,
		Code: invite.Code,
	}

	NewResponse(c, http.StatusCreated, output)
}

type invitationGetAllResponse []invitationResponseModel

// @Summary Получить все приглашения организации
// @Success 200 {object} invitationGetAllResponse "список инвайтов"
// @Router /invites [get]
func (s *invitation) GetAll(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	where := &repository.InvitationModel{
		OrgID: claims.OrganizationID,
	}

	invites, err := s.repo.Invitation.Find(where)
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}

	output := make(invitationGetAllResponse, len(*invites))

	for i, invite := range *invites {
		output[i] = invitationResponseModel{
			ID:             invite.ID,
			AffiliateOrgID: invite.AffiliateOrgID,

			Code:      invite.Code,
			ExpiresIn: invite.ExpiresIn,
		}
	}

	NewResponse(c, http.StatusCreated, output)
}

// @Summary Удалить приглашение
// @Success 200 {object} object "object"
// @Router /invites/:id [delete]
func (s *invitation) Delete(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	inviteID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	where := &repository.InvitationModel{
		ID:    uint(inviteID),
		OrgID: claims.OrganizationID,
	}

	if err := s.repo.Invitation.Delete(where); err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

type invitationGetNotActivatedResponse struct {
	ID        uint   `json:"id"`
	Code      string `json:"code"`
	ExpiresIn int64  `json:"expires_in"`
}

// @Summary Получить неактивированные приглашения организации
// @Accept json
// @Produce json
// @Success 200 {object} []invitationGetNotActivatedResponse "object"
// @Router /invites.NotActivated [get]
func (s *invitation) GetNotActivated(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	where := &repository.InvitationModel{
		OrgID: claims.OrganizationID,
	}

	invites, err := s.repo.Invitation.FindNotActivated(where)
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}

	output := make([]invitationGetNotActivatedResponse, len(*invites))
	for i, invite := range *invites {
		output[i] = invitationGetNotActivatedResponse{
			ID:        invite.ID,
			Code:      invite.Code,
			ExpiresIn: invite.ExpiresIn,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

type invitationGetActivatedAffiliateOrgOutlet struct {
	ID   uint   `json:"id"`   // id точки
	Name string `json:"name"` // имя точки
}

type invitationGetActivateAffiliateOrg struct {
	ID      uint                                       `json:"id"`      //id филиала
	Name    string                                     `json:"name"`    //имя филиала
	Outlets []invitationGetActivatedAffiliateOrgOutlet `json:"outlets"` //точки филиала
}

type invitationGetActivatedResponse struct {
	ID           uint                              `json:"id"`            //id инвайта
	AffiliateOrg invitationGetActivateAffiliateOrg `json:"affiliate_org"` //информация о филиале
}

// @Summary Получить активированные приглашения организации
// @Accept json
// @Produce json
// @Success 200 {object} []InvitationGetActivatedOutput "возвращамый объект"
// @Router /invites.Activated [get]
func (s *invitation) GetActivated(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	where := &repository.InvitationModel{
		OrgID: claims.OrganizationID,
	}

	invites, err := s.repo.Invitation.FindActivated(where)
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
		return
	}

	response := make([]invitationGetActivatedResponse, len(*invites))

	for i, invite := range *invites {
		org, err := s.repo.Organizations.FindFirts(&repository.OrganizationModel{ID: invite.AffiliateOrgID})
		if err != nil {
			NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
			return
		}

		outlets, err := s.repo.Outlets.Find(&repository.OutletModel{OrgID: org.ID})
		if err != nil {
			NewResponse(c, http.StatusBadRequest, errUnknown(err.Error()))
			return
		}

		affiliateOrg := invitationGetActivateAffiliateOrg{
			ID:      org.ID,
			Name:    org.Name,
			Outlets: make([]invitationGetActivatedAffiliateOrgOutlet, len(*outlets)),
		}

		for j, outlet := range *outlets {
			affiliateOrg.Outlets[j] = invitationGetActivatedAffiliateOrgOutlet{
				ID:   outlet.ID,
				Name: outlet.Name,
			}
		}

		response[i] = invitationGetActivatedResponse{
			ID:           invite.ID,
			AffiliateOrg: affiliateOrg,
		}
	}

	NewResponse(c, http.StatusOK, response)
}
