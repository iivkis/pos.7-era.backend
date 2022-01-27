package repository

import "gorm.io/gorm"

type OutletsRepository interface {
	Create(m *OutletModel) error
	GetAll(orgID uint) ([]OutletModel, error)
}

type OutletModel struct {
	gorm.Model
	OrgID uint

	Name string

	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type outlets struct {
	db *gorm.DB
}

func newOutletsRepo(db *gorm.DB) *outlets {
	return &outlets{
		db: db,
	}
}

func (r *outlets) Create(m *OutletModel) error {
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	return nil
}

func (r *outlets) GetAll(orgID uint) ([]OutletModel, error) {
	var models []OutletModel
	if err := r.db.Where("org_id = ?", orgID).Find(&models).Error; err != nil {
		return models, err
	}
	return models, nil
}
