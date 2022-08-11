package repository

import "gorm.io/gorm"

type OrderListModel struct {
	gorm.Model

	ProductID   uint
	OrderInfoID uint
	SessionID   uint
	OutletID    uint
	OrgID       uint

	ProductName  string
	ProductPrice float64
	Count        int

	ProductModel      ProductModel      `gorm:"foreignKey:ProductID"`
	OrderInfoModel    OrderInfoModel    `gorm:"foreignKey:OrderInfoID"`
	SessionModel      SessionModel      `gorm:"foreignKey:SessionID"`
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

func (r *OrderListRepo) Create(m *OrderListModel) (err error) {
	return r.db.Create(m).Error
}

func (r *OrderListRepo) Find(where *OrderListModel) (result *[]OrderListModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r *OrderListRepo) FindUnscoped(where *OrderListModel) (result *[]OrderListModel, err error) {
	err = r.db.Unscoped().Where(where).Find(&result).Error
	return
}

func (r *OrderListRepo) FindForCalculation(where *OrderListModel) (result *[]OrderListModel, err error) {
	err = r.db.Select("product_price, count").
		Where(where).
		Find(&result).
		Error
	return
}

func (r *OrderListRepo) Updates(where *OrderListModel, updates *OrderListModel) error {
	return r.db.Where(where).Updates(updates).Error
}

func (r *OrderListRepo) Delete(where *OrderListModel) (err error) {
	err = r.db.Where(where).Delete(&OrderListModel{}).Error
	return
}

func (r *OrderListRepo) Recovery(where *OrderListModel) (err error) {
	err = r.db.Model(&OrderListModel{}).Unscoped().Where(where).UpdateColumn("deleted_at", nil).Error
	return
}
