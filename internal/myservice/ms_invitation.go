package myservice

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type InvitationOutputModel struct {
	ID uint `json:"id"` //id инвайта

	Code      string `json:"code"`       // рандомный код (равен "", если активирован)
	ExpiresIn int64  `json:"expires_in"` // во сколько истекает инвайт(равен 0, если активирован)

	AffiliateOrgID uint `json:"affiliate_org_id"` // приглашенная организация
}

type InvitationService struct {
	repo *repository.Repository
}

func newInvitationService(repo *repository.Repository) *InvitationService {
	return &InvitationService{
		repo: repo,
	}
}

type InvitationCreateOutput struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
}

//@Summary Создать код приглашения
//@Accept json
//@Produce json
//@Success 201 {object} InvitationCreateOutput "возвращает id созданной записи и код приглашения"
//@Failure 400 {object} serviceError
//@Router /invites [post]
func (s *InvitationService) Create(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	//запретить, если оргнизация сама является дочерней
	if s.repo.Invitation.Exists(&repository.InvitationModel{AffiliateOrgID: claims.OrganizationID}) {
		NewResponse(c, http.StatusBadRequest, errPermissionDenided("an affiliated organization cannot create invitation codes"))
		return
	}

	if n, err := s.repo.Invitation.CountNotActivated(claims.OrganizationID); err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
		return
	} else if n >= 10 {
		NewResponse(c, http.StatusBadRequest, errRecordAlreadyExists("you can create up to 10 inactive invites"))
		return
	}

	invite := &repository.InvitationModel{
		OrgID: claims.OrganizationID,
	}

	if err := s.repo.Invitation.Create(invite); err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
		return
	}

	output := InvitationCreateOutput{
		ID:   invite.ID,
		Code: invite.Code,
	}

	NewResponse(c, http.StatusCreated, output)
}

type InvitationGetAllOutput []InvitationOutputModel

//@Summary Получить все приглашения организации
//@Accept json
//@Produce json
//@Success 200 {object} InvitationGetAllOutput "возвращамый объект"
//@Failure 400 {object} serviceError
//@Router /invites [get]
func (s *InvitationService) GetAll(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	where := &repository.InvitationModel{
		OrgID: claims.OrganizationID,
	}

	invites, err := s.repo.Invitation.Find(where)
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
		return
	}

	output := make(InvitationGetAllOutput, len(*invites))
	for i, invite := range *invites {
		output[i] = InvitationOutputModel{
			ID:             invite.ID,
			Code:           invite.Code,
			ExpiresIn:      invite.ExpiresIn,
			AffiliateOrgID: invite.AffiliateOrgID,
		}
	}

	NewResponse(c, http.StatusCreated, output)
}

//@Summary Удалить приглашение
//@Accept json
//@Produce json
//@Success 200 {object} object "возвращамый объект"
//@Failure 400 {object} serviceError
//@Router /invites/:id [delete]
func (s *InvitationService) Delete(c *gin.Context) {
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
		NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

type InvitationGetNotActivatedOutput struct {
	ID        uint   `json:"id"`         //id инвайта
	Code      string `json:"code"`       // рандомный код (равен "", если активирован)
	ExpiresIn int64  `json:"expires_in"` // во сколько истекает инвайт(равен 0, если активирован)
}

//@Summary Получить неактивированные приглашения организации
//@Accept json
//@Produce json
//@Success 200 {object} []InvitationGetNotActivatedOutput "возвращамый объект"
//@Failure 400 {object} serviceError
//@Router /invites.NotActivated [get]
func (s *InvitationService) GetNotActivated(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	where := &repository.InvitationModel{
		OrgID: claims.OrganizationID,
	}

	invites, err := s.repo.Invitation.FindNotActivated(where)
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
		return
	}

	output := make([]InvitationGetNotActivatedOutput, len(*invites))
	for i, invite := range *invites {
		output[i] = InvitationGetNotActivatedOutput{
			ID:        invite.ID,
			Code:      invite.Code,
			ExpiresIn: invite.ExpiresIn,
		}
	}

	NewResponse(c, http.StatusOK, output)
}

type InvitationGetActivatedFieldOutlets struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type InvitationGetActivatedFieldAffiliateOrg struct {
	ID      uint                                 `json:"id"`
	Name    string                               `json:"name"`
	Outlets []InvitationGetActivatedFieldOutlets `json:"outlets"`
}

type InvitationGetActivatedOutput struct {
	ID           uint                                    `json:"id"` //id инвайта
	AffiliateOrg InvitationGetActivatedFieldAffiliateOrg `json:"affiliate_org"`
}

//@Summary Получить активированные приглашения организации
//@Accept json
//@Produce json
//@Success 200 {object} []InvitationGetActivatedOutput "возвращамый объект"
//@Failure 400 {object} serviceError
//@Router /invites.NotActivated [get]
func (s *InvitationService) GetActivated(c *gin.Context) {
	claims := mustGetEmployeeClaims(c)

	where := &repository.InvitationModel{
		OrgID: claims.OrganizationID,
	}

	invites, err := s.repo.Invitation.FindActivated(where)
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
		return
	}

	output := make([]InvitationGetActivatedOutput, len(*invites))
	for i, invite := range *invites {
		org, err := s.repo.Organizations.FindFirts(&repository.OrganizationModel{ID: invite.AffiliateOrgID})
		if err != nil {
			NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
			return
		}

		outlets, err := s.repo.Outlets.Find(&repository.OutletModel{OrgID: invite.OrgID})
		if err != nil {
			NewResponse(c, http.StatusBadRequest, errUnknownDatabase(err.Error()))
			return
		}

		affiliateOrg := InvitationGetActivatedFieldAffiliateOrg{
			ID:      org.ID,
			Name:    org.Name,
			Outlets: make([]InvitationGetActivatedFieldOutlets, len(*outlets)),
		}

		for i, outlet := range *outlets {
			affiliateOrg.Outlets[i] = InvitationGetActivatedFieldOutlets{
				ID:   outlet.ID,
				Name: outlet.Name,
			}
		}

		output[i] = InvitationGetActivatedOutput{
			ID:           invite.ID,
			AffiliateOrg: affiliateOrg,
		}
	}

	NewResponse(c, http.StatusOK, output)
}
