package repository

import "gorm.io/gorm"

type OrderInfoModel struct {
	gorm.Model

	SessionID uint
	OutletID  uint
	OrgID     uint

	Date int64

	PayType      int // 0 - наличные, 1 - безналичные, 2 - смешанный
	EmployeeName string

	SessionModel      SessionModel      `gorm:"foreignKey:SessionID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type OrderInfoRepo struct {
	db *gorm.DB
}

func newOrderInfoRepo(db *gorm.DB) *OrderInfoRepo {
	return &OrderInfoRepo{
		db: db,
	}
}

//actual
func (r *OrderInfoRepo) Create(model *OrderInfoModel) error {
	return r.db.Create(model).Error
}

func (r OrderInfoRepo) Find(where *OrderInfoModel) (result *[]OrderInfoModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r OrderInfoRepo) FindUnscoped(where *OrderInfoModel) (result *[]OrderInfoModel, err error) {
	err = r.db.Unscoped().Where(where).Find(&result).Error
	return
}

func (r OrderInfoRepo) FindFirst(where *OrderInfoModel) (result *OrderInfoModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r OrderInfoRepo) FindFirstUnscoped(where *OrderInfoModel) (result *OrderInfoModel, err error) {
	err = r.db.Unscoped().Where(where).Find(&result).Error
	return
}

func (r *OrderInfoRepo) Updates(where *OrderInfoModel, updatedFields *OrderInfoModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *OrderInfoRepo) Delete(where *OrderInfoModel) (err error) {
	err = r.db.Where(where).Delete(&OrderInfoModel{}).Error
	return
}

func (r *OrderInfoRepo) Exists(where *OrderInfoModel) bool {
	return r.db.Select("id").Where(where).First(&OrderInfoModel{}).Error == nil
}

func (r *OrderInfoRepo) ExistsUnscoped(where *OrderInfoModel) bool {
	return r.db.Unscoped().Select("id").Where(where).First(&OrderInfoModel{}).Error == nil
}

func (r *OrderInfoRepo) Recovery(where *OrderInfoModel) (err error) {
	err = r.db.Model(&OrderInfoModel{}).Unscoped().Where(where).UpdateColumn("deleted_at", nil).Error
	return
}

func (r *OrderInfoRepo) Count(where *OrderInfoModel) (n int64, err error) {
	err = r.db.Model(where).Where(where).Count(&n).Error
	return
}
