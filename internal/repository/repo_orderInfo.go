package repository

import "gorm.io/gorm"

type OrderInfoModel struct {
	gorm.Model

	PayType      int // 0 - наличные, 1 - безналичные, 2 - смешанный
	Date         int64
	EmployeeName string
	SessionID    uint

	OrgID    uint
	OutletID uint

	SessionModel SessionModel `gorm:"foreignKey:SessionID"`

	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
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

func (r *OrderInfoRepo) Updates(where *OrderInfoModel, updatedFields *OrderInfoModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *OrderInfoRepo) Delete(where *OrderInfoModel) (err error) {
	err = r.db.Where(where).Delete(&OrderInfoModel{}).Error
	return
}

func (r *OrderInfoRepo) Exists(where *OrderInfoModel) bool {
	return r.db.Where(where).First(&OrderInfoModel{}).Error == nil
}
