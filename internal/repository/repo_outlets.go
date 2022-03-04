package repository

import "gorm.io/gorm"

type OutletModel struct {
	gorm.Model

	Name  string
	OrgID uint

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
	return r.db.Create(m).Error
}

func (r *OutletsRepo) Updates(m *OutletModel, outletID interface{}) error {
	return r.db.Where("id = ?", outletID).Updates(m).Error
}

func (r *OutletsRepo) Delete(outletID interface{}) error {
	return r.db.Where("id = ?", outletID).Delete(&OutletModel{}).Error
}

func (r *OutletsRepo) FindAllByOrgID(orgID interface{}) (models []OutletModel, err error) {
	err = r.db.Where("org_id = ?", orgID).Find(&models).Error
	return
}

func (r *OutletsRepo) ExistsInOrg(outletID interface{}, orgID interface{}) bool {
	err := r.db.Where("id = ? AND org_id = ?", outletID, orgID).First(&OutletModel{}).Error
	return err == nil
}
