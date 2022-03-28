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

	AffiliateOrgID uint // приглашенная организация
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

	if n, err := s.repo.Invitation.CountNotActived(claims.OrganizationID); err != nil {
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
