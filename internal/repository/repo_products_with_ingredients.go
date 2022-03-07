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
	return r.db.Where("id = ? AND outlet_id = ?", ID, outletID).Delete(&ProductWithIngredientModel{}).Error
}

func (r *ProductsWithIngredientsRepo) FindAllByOrgID(orgID interface{}) (m []ProductWithIngredientModel, err error) {
	err = r.db.Where("org_id = ?", orgID).Find(&m).Error
	return
}

func (r *ProductsWithIngredientsRepo) FindAllByOutletID(outletID interface{}, whereProductID uint) (m []ProductWithIngredientModel, err error) {
	if whereProductID == 0 {
		err = r.db.Where("outlet_id = ? ", outletID).Find(&m).Error
	} else {
		err = r.db.Where("outlet_id = ? AND product_id = ?", outletID, whereProductID).Find(&m).Error
	}
	return
}

// func (r *ProductsWithIngredientsRepo) FindAllByProductID(productID interface{}, inOutletID interface{}) (m []ProductWithIngredientModel, err error) {
// 	err = r.db.Where("product_id = ? AND outlet_id = ?", productID, inOutletID).Find(&m).Error
// 	return
// }

func (r *ProductsWithIngredientsRepo) WriteOffIngredients(productID interface{}, count int, inOutletID interface{}) (err error) {
	var list []ProductWithIngredientModel
	if err = r.db.Where("product_id = ? AND outlet_id = ?", productID, inOutletID).Find(&list).Error; err != nil {
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
