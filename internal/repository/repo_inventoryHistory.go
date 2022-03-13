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

func (r *InventoryHistoryRepo) Create(m *InventoryHistoryModel) error {
	return r.db.Create(m).Error
}

func (r InventoryHistoryRepo) Find(where *InventoryHistoryModel) (result *[]InventoryHistoryModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r InventoryHistoryRepo) FindWithPeriod(where *InventoryHistoryModel, start uint64, end uint64) (result *[]InventoryHistoryModel, err error) {
	if end == 0 {
		err = r.db.Where("date >= ?", start).Find(&result, where).Error
	} else {
		err = r.db.Where("date >= ? AND date <= ?", start, end).Find(&result, where).Error
	}
	return
}

func (r *InventoryHistoryRepo) Updates(where *InventoryHistoryModel, updatedFields *InventoryHistoryModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *InventoryHistoryRepo) Delete(where *InventoryHistoryModel) (err error) {
	err = r.db.Where(where).Delete(&InventoryHistoryModel{}).Error
	return
}

func (r *InventoryHistoryRepo) Exists(where *InventoryHistoryModel) bool {
	return r.db.Where(where).First(&InventoryHistoryModel{}).Error == nil
}
