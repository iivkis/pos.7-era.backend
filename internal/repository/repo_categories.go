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

//actual
func (r *CategoriesRepo) Create(m *CategoryModel) (err error) {
	return r.db.Create(m).Error
}

func (r *CategoriesRepo) Find(where *CategoryModel) (result *[]CategoryModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r *CategoriesRepo) Updates(where *CategoryModel, updatedFields *CategoryModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *CategoriesRepo) Delete(where *CategoryModel) (err error) {
	err = r.db.Where(where).Delete(&CategoryModel{}).Error
	return
}

func (r *CategoriesRepo) Exists(where *CategoryModel) bool {
	return r.db.Where(where).First(&CategoryModel{}).Error == nil
}
