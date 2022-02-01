package repository

import "gorm.io/gorm"

type OutletModel struct {
	gorm.Model
	OrgID uint

	Name string

	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type OutletsRepo struct {
	db *gorm.DB
}

func newOutletsRepo(db *gorm.DB) *OutletsRepo {
	return &OutletsRepo{
		db: db,
	}
}

func (r *OutletsRepo) Create(m *OutletModel) error {
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	return nil
}

func (r *OutletsRepo) GetAll(orgID uint) ([]OutletModel, error) {
	var models []OutletModel
	if err := r.db.Where("org_id = ?", orgID).Find(&models).Error; err != nil {
		return models, err
	}
	return models, nil
}
