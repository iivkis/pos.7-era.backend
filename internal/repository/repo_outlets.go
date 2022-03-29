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

//actual
func (r *OutletsRepo) Create(m *OutletModel) error {
	return r.db.Create(m).Error
}

func (r *OutletsRepo) Updates(where *OutletModel, updatedFields *OutletModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *OutletsRepo) Delete(where *OutletModel) error {
	return r.db.Where(where).Delete(&OutletModel{}).Error
}

func (r *OutletsRepo) Find(where *OutletModel) (result *[]OutletModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r *OutletsRepo) Exists(where *OutletModel) bool {
	return r.db.Select("id").Where(where).First(&OutletModel{}).Error == nil
}

func (r *OutletsRepo) ExistsInOrg(outletID uint, orgID uint) bool {
	return r.db.Select("id").Where("id = ? AND org_id = ?", outletID, orgID).First(&OutletModel{}).Error == nil
}
