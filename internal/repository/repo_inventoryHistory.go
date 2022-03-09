package repository

import "gorm.io/gorm"

type InventoryHistoryModel struct {
	gorm.Model

	Date       int64 //unixmilli
	EmployeeID uint  //сотрудник, который делал инветаризацию
	OutletID   uint
	OrgID      uint

	EmployeeModel     EmployeeModel     `gorm:"foreignKey:EmployeeID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type InventoryHistoryRepo struct {
	db *gorm.DB
}

func newInventoryHistoryRepo(db *gorm.DB) *InventoryHistoryRepo {
	return &InventoryHistoryRepo{
		db: db,
	}
}

func (r *InventoryHistoryRepo) Create(m *IngredientModel) error {
	return r.db.Create(m).Error
}

func (r *InventoryHistoryRepo) FindAllByOutletID(outletID interface{}) (m []IngredientModel, err error) {
	err = r.db.Where("outlet_id = ?", outletID).Find(&m).Error
	return
}
