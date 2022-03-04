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
func (r *OrderInfoRepo) Create(m *OrderInfoModel) error {
	return r.db.Create(m).Error
}

func (r *OrderInfoRepo) Updates(m *OrderInfoModel, orderInfoID interface{}, outletID interface{}) error {
	return r.db.Where("id = ? AND outlet_id = ?", orderInfoID, outletID).Updates(m).Error
}

func (r *OrderInfoRepo) FindAllByOrgID(orgID interface{}) (m []OrderInfoModel, err error) {
	err = r.db.Unscoped().Where("org_id = ?", orgID).Find(&m).Error
	return
}

func (r *OrderInfoRepo) FindAllByOutletID(outletID interface{}) (m []OrderInfoModel, err error) {
	err = r.db.Unscoped().Where("outlet_id = ?", outletID).Find(&m).Error
	return
}

func (r *OrderInfoRepo) Delete(orderInfoID interface{}, outletID interface{}) error {
	return r.db.Where("id = ? AND outlet_id = ?", orderInfoID, outletID).Error
}
