package repository

import "gorm.io/gorm"

type CategoryModel struct {
	gorm.Model
	Name string

	OutletID    uint
	OutletModel OutletModel `gorm:"foreignKey:OutletID"`
}

type CategoryRepository interface {
	Create(m *CategoryModel) error
	GetAllByOutletID(outletID uint) (cats []CategoryModel, err error)
}

type category struct {
	db *gorm.DB
}

func newCategoryRepo(db *gorm.DB) *category {
	return &category{
		db: db,
	}
}

func (r *category) Create(m *CategoryModel) (err error) {
	err = r.db.Create(m).Error
	return
}

func (r *category) GetAllByOutletID(outletID uint) (cats []CategoryModel, err error) {
	err = r.db.Where("outlet_id = ?", outletID).Find(&cats).Error
	return
}

func (r *category) DeleteByID(outletID uint, categoryID uint) (err error) {
	err = r.db.Where("id = ? AND outlet_id = ?", categoryID, outletID).Delete(&CategoryModel{}).Error
	return
}