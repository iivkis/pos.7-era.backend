package repository

import "gorm.io/gorm"

type iOrganizations interface {
	Create(orgModel *OrganizationModel) error
	GetAll(orgsModel *[]OrganizationModel) error
}

type organizationsRepo struct {
	db *gorm.DB
}

type OrganizationModel struct {
	gorm.Model
	Name    string
	Email   string
	Pwdhash string
}

func newOrganizationsRepo(db *gorm.DB) *organizationsRepo {
	return &organizationsRepo{db: db}
}

func (r *organizationsRepo) Create(orgModel *OrganizationModel) error {
	return r.db.Create(orgModel).Error
}

func (r *organizationsRepo) GetAll(orgsModel *[]OrganizationModel) error {
	return r.db.Find(orgsModel).Error
}
