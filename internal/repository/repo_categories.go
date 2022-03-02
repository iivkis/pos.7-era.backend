package repository

import "gorm.io/gorm"

type CategoryModel struct {
	gorm.Model
	Name string

	OutletID uint
	OrgID    uint

	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
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

func (r *CategoriesRepo) FindAllByOutletID(outletID interface{}) (cats []CategoryModel, err error) {
	err = r.db.Where("outlet_id = ?", outletID).Find(&cats).Error
	return
}

func (r *CategoriesRepo) FindAllByOrgID(orgID uint) (cats []CategoryModel, err error) {
	err = r.db.Where("org_id = ?", orgID).Find(&cats).Error
	return
}

func (r *CategoriesRepo) DeleteByID(outletID interface{}, paramCategoryID string) (err error) {
	err = r.db.Where("id = ? AND outlet_id = ?", paramCategoryID, outletID).
		Delete(&CategoryModel{}).Error
	return
}

func (r *CategoriesRepo) Updates(categoryID interface{}, outletID interface{}, m *CategoryModel) (err error) {
	err = r.db.Where("id = ? AND outlet_id = ?", categoryID, outletID).Updates(m).Error
	return err
}
