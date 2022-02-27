package repository

import "gorm.io/gorm"

type ProductWithIngredientModel struct {
	gorm.Model
	CountTakeForSell float64

	ProductID    uint
	IngredientID uint
	OutletID     uint
	OrgID        uint

	ProductModel      ProductModel      `gorm:"foreignKey:ProductID"`
	IngredientModel   IngredientModel   `gorm:"foreignKey:IngredientID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type ProductsWithIngredientsRepo struct {
	db *gorm.DB
}

func newProductsWithIngredientsRepo(db *gorm.DB) *ProductsWithIngredientsRepo {
	return &ProductsWithIngredientsRepo{
		db: db,
	}
}

func (r *ProductsWithIngredientsRepo) Create(m *ProductWithIngredientModel) error {
	return r.db.Create(m).Error
}

func (r *ProductsWithIngredientsRepo) Updates(m *ProductWithIngredientModel, ID interface{}, outletID interface{}) error {
	return r.db.Model(&ProductWithIngredientModel{}).
		Where("id = ? AND outlet_id = ?", ID, outletID).
		Updates(m).Error
}

func (r *ProductsWithIngredientsRepo) Delete(ID interface{}, outletID interface{}) error {
	return r.db.Where("id = ? AND outlet_id = ?", ID, outletID).Delete(&ProductModel{}).Error
}

func (r *ProductsWithIngredientsRepo) FindAllByOrgID(orgID interface{}) (m []ProductWithIngredientModel, err error) {
	err = r.db.Where("org_id = ?", orgID).Find(&m).Error
	return
}

func (r *ProductsWithIngredientsRepo) FindAllByOutletID(outlet interface{}) (m []ProductWithIngredientModel, err error) {
	err = r.db.Where("outlet_id = ?", outlet).Find(&m).Error
	return
}

func (r *ProductsWithIngredientsRepo) FindAllByProductID(productID interface{}) (m []ProductWithIngredientModel, err error) {
	err = r.db.Where("product_id = ?", productID).Find(&m).Error
	return
}

func (r *ProductsWithIngredientsRepo) WriteOffIngredients(productID interface{}, count int, outletID interface{}) (err error) {
	var list []ProductWithIngredientModel
	if err = r.db.Where("product_id = ? AND outlet_id = ?", productID, outletID).Find(&list).Error; err != nil {
		return err
	}

	for _, item := range list {
		var m IngredientModel
		if err = r.db.Where("id = ?", item.IngredientID).First(&m).Error; err != nil {
			return err
		}

		//write off
		m.Count -= item.CountTakeForSell * float64(count)

		if err = r.db.Updates(&m).Error; err != nil {
			return err
		}
	}
	return
}
