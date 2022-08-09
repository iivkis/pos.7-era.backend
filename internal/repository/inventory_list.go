package repository

import "gorm.io/gorm"

type InventoryListModel struct {
	gorm.Model

	OldCount  float64 //значение остатков, которое имеется в системе
	NewCount  float64 //новое значение по факту, вводит кассир
	LossPrice float64 //сумма, которую потеряла точка, если кол-во товаров в системе и фактическое кол-во товаров не совпадают

	IngredientID       uint
	InventoryHistoryID uint
	OutletID           uint
	OrgID              uint

	IngredientModel       IngredientModel       `gorm:"foreignKey:IngredientID"`
	InventoryHistoryModel InventoryHistoryModel `gorm:"foreignKey:InventoryHistoryID"`
	OutletModel           OutletModel           `gorm:"foreignKey:OutletID"`
	OrganizationModel     OrganizationModel     `gorm:"foreignKey:OrgID"`
}

type InventoryListRepo struct {
	db *gorm.DB
}

func newInventoryListRepo(db *gorm.DB) *InventoryListRepo {
	return &InventoryListRepo{
		db: db,
	}
}

func (r *InventoryListRepo) Create(m *InventoryListModel) error {
	return r.db.Create(m).Error
}

func (r InventoryListRepo) Find(where *InventoryListModel) (result *[]InventoryListModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r InventoryListRepo) FindFirts(where *InventoryListModel) (result *InventoryListModel, err error) {
	err = r.db.Where(where).First(&result).Error
	return
}

func (r *InventoryListRepo) Updates(where *InventoryListModel, updatedFields *InventoryListModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *InventoryListRepo) Delete(where *InventoryListModel) (err error) {
	err = r.db.Where(where).Delete(&InventoryListModel{}).Error
	return
}

func (r *InventoryListRepo) Exists(where *InventoryListModel) bool {
	return r.db.Select("id").Where(where).First(&InventoryListModel{}).Error == nil
}
