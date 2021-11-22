package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
)

type iOrganizations interface {
	GetAllOrganizations(c *gin.Context)
	AddOrganization(c *gin.Context)
}

type organizations struct {
	repo repository.Repository
}

type organizationInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type organizationOutput struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func newOrganizations(repo repository.Repository) *organizations {
	return &organizations{repo: repo}
}

func (s *organizations) GetAllOrganizations(c *gin.Context) {
	var orgsModel []repository.OrganizationModel
	if err := s.repo.Organizations.GetAll(&orgsModel); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	orgsOutput := make([]organizationOutput, len(orgsModel))
	for i, org := range orgsModel {
		orgsOutput[i] = organizationOutput{
			ID:    org.ID,
			Name:  org.Name,
			Email: org.Email,
		}
	}

	c.JSON(http.StatusOK, orgsOutput)
}

func (s *organizations) AddOrganization(c *gin.Context) {
	var orgInput organizationInput
	if err := c.ShouldBindJSON(&orgInput); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	var orgModel repository.OrganizationModel
	orgModel.Name = orgInput.Name
	orgModel.Email = orgInput.Email
	orgModel.Pwdhash = orgInput.Password // !TODO: Need hash password

	if err := s.repo.Organizations.Create(&orgModel); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, organizationOutput{
		ID:    orgModel.ID,
		Name:  orgModel.Name,
		Email: orgModel.Email,
	})
}
