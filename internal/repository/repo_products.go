package repository

import "gorm.io/gorm"

type ProductModel struct {
	gorm.Model

	Name   string
	Amount int
	Price  float64
	Photo  string

	CategoryID uint
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

//Возвращает все продукты текущей точки
func (r *ProductsRepo) GetAllForOutlet(outletID interface{}) (products []ProductModel, err error) {
	err = r.db.Where("outlet_id = ?", outletID).Find(&products).Error
	return
}

func (r *ProductsRepo) GetAllForOrg(orgID interface{}) (products []ProductModel, err error) {
	err = r.db.Where("org_id = ?", orgID).Find(&products).Error
	return
}

func (r *ProductsRepo) GetOneForOutlet(productID interface{}, outletID interface{}) (product ProductModel, err error) {
	err = r.db.Where("id = ? AND outlet_id = ?", productID, outletID).First(&product).Error
	return
}

func (r *ProductsRepo) Create(product *ProductModel) (err error) {
	err = r.db.Create(product).Error
	return err
}

func (r *ProductsRepo) Update(productID interface{}, outletID interface{}, product *ProductModel) (err error) {
	err = r.db.Model(&ProductModel{}).Where("id = ? AND outlet_id = ?", productID, outletID).Updates(product).Error
	return
}

func (r *ProductsRepo) Delete(productID interface{}, outletID interface{}) (err error) {
	err = r.db.Model(&ProductModel{}).Delete("id = ? AND outlet_id = ?", productID, outletID).Error
	return
}
