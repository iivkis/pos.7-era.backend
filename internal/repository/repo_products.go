package repository

import "gorm.io/gorm"

type ProductModel struct {
	ID uint

	Name           string
	ProductNameKKT string
	Barcode        int

	Amount int
	Price  float64
	Photo  string

	CategoryID uint `gorm:"default:NULL"`
	OutletID   uint
	OrgID      uint

	CategoryModel     CategoryModel     `gorm:"foreignKey:CategoryID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type ProductsRepo struct {
	db *gorm.DB
}

func newProductsRepo(db *gorm.DB) *ProductsRepo {
	return &ProductsRepo{
		db: db,
	}
}

//actual
func (r *ProductsRepo) Create(product *ProductModel) error {
	return r.db.Create(product).Error
}

func (r *ProductsRepo) Find(where *ProductModel) (result *[]ProductModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r *ProductsRepo) FindFirst(where *ProductModel) (result *ProductModel, err error) {
	err = r.db.Where(where).First(&result).Error
	return
}

func (r *ProductsRepo) Updates(where *ProductModel, updatedFields *ProductModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *ProductsRepo) Delete(where *ProductModel) (err error) {
	err = r.db.Where(where).Delete(&ProductModel{}).Error
	return
}

func (r *ProductsRepo) Exists(where *ProductModel) bool {
	return r.db.Select("id").Where(where).First(&ProductModel{}).Error == nil
}
