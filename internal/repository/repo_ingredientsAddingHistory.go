package repository

import "gorm.io/gorm"

type IngredientsAddingHistoryModel struct {
	ID uint

	Count  int     //кол-во продукта, который не сходится
	Total  float64 //сумма, на которую не сходится
	Status int     // 1 - инвенторизация

	Date int64 //unixmilli

	IngredientID uint
	EmployeeID   uint //сотрудник, который делал инветаризацию
	OutletID     uint
	OrgID        uint

	IngredientModel   IngredientModel   `gorm:"foreignKey:IngredientID"`
	EmployeeModel     EmployeeModel     `gorm:"foreignKey:EmployeeID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type IngredientsAddingHistoryRepo struct {
	db *gorm.DB
}

func newIngredientsAddingHistoryRepo(db *gorm.DB) *IngredientsAddingHistoryRepo {
	return &IngredientsAddingHistoryRepo{
		db: db,
	}
}

func (r *IngredientsAddingHistoryRepo) Create(m *IngredientsAddingHistoryModel) error {
	return r.db.Create(m).Error
}

func (r IngredientsAddingHistoryRepo) Find(where *IngredientsAddingHistoryModel) (result *[]IngredientsAddingHistoryModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r IngredientsAddingHistoryRepo) FindWithPeriod(where *IngredientsAddingHistoryModel, start uint64, end uint64) (result *[]IngredientsAddingHistoryModel, err error) {
	if end == 0 {
		err = r.db.Where("date >= ?", start).Find(&result, where).Error
	} else {
		err = r.db.Where("date >= ? AND date <= ?", start, end).Find(&result, where).Error
	}
	return
}

func (r *IngredientsAddingHistoryRepo) Updates(where *IngredientsAddingHistoryModel, updatedFields *IngredientsAddingHistoryModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *IngredientsAddingHistoryRepo) Delete(where *IngredientsAddingHistoryModel) (err error) {
	err = r.db.Where(where).Delete(&InventoryHistoryModel{}).Error
	return
}

func (r *IngredientsAddingHistoryRepo) Exists(where *IngredientsAddingHistoryModel) bool {
	return r.db.Where(where).First(&IngredientsAddingHistoryModel{}).Error == nil
}
