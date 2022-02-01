package repository

import "gorm.io/gorm"

type CategoryModel struct {
	gorm.Model
	Name string

	OutletID    uint
	OutletModel OutletModel `gorm:"foreignKey:OutletID"`
}

type CategoriesRepo struct {
	db *gorm.DB
}

func newCategoriesRepo(db *gorm.DB) *CategoriesRepo {
	return &CategoriesRepo{
		db: db,
	}
}

func (r *CategoriesRepo) Create(m *CategoryModel) (err error) {
	err = r.db.Create(m).Error
	return
}

func (r *CategoriesRepo) GetAllByOutletID(outletID uint) (cats []CategoryModel, err error) {
	err = r.db.Where("outlet_id = ?", outletID).Find(&cats).Error
	return
}

func (r *CategoriesRepo) DeleteByID(outletID uint, paramCategoryID string) (err error) {
	err = r.db.Where("id = ? AND outlet_id = ?", paramCategoryID, outletID).
		Delete(&CategoryModel{}).Error
	return
}

func (r *CategoriesRepo) Update(paramCategoryID string, m *CategoryModel) (err error) {
	err = r.db.Where("id = ?", paramCategoryID).Updates(m).Error
	return err
}
