package repository

import "gorm.io/gorm"

type OrderListModel struct {
	gorm.Model
	ProductName string

	ProductPrice float64
	Count        int

	ProductID   uint
	OrderInfoID uint

	OutletID uint
	OrgID    uint

	ProductModel   ProductModel   `gorm:"foreignKey:ProductID"`
	OrderInfoModel OrderInfoModel `gorm:"foreignKey:OrderInfoID"`

	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type OrderListRepo struct {
	db *gorm.DB
}

func newOrderListRepo(db *gorm.DB) *OrderListRepo {
	return &OrderListRepo{
		db: db,
	}
}

func (r *OrderListRepo) Create(m *OrderListModel) error {
	return r.db.Create(m).Error
}

func (r *OrderListRepo) FindAllForOrg(orgID interface{}) (orders []OrderListModel, err error) {
	err = r.db.Where("org_id = ?", orgID).Find(&orders).Error
	return
}

func (r *OrderListRepo) Updates(m *OrderListModel, orderListID interface{}, outletID interface{}) error {
	return r.db.Where("id = ? AND outlet_id = ?", orderListID, outletID).Updates(m).Error
}

func (r *OrderListRepo) Delete(orderListID interface{}, outletID interface{}) error {
	return r.db.Where("id = ? AND outlet_id = ?", orderListID, outletID).Delete(&OrderListModel{}).Error
}
