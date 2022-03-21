package repository

import (
	"gorm.io/gorm"
)

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

//actual
func (r *ProductsWithIngredientsRepo) Create(m *ProductWithIngredientModel) error {
	return r.db.Create(m).Error
}

func (r *ProductsWithIngredientsRepo) Find(where *ProductWithIngredientModel) (result *[]ProductWithIngredientModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r *ProductsWithIngredientsRepo) Updates(where *ProductWithIngredientModel, updatedFields *ProductWithIngredientModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *ProductsWithIngredientsRepo) Delete(where *ProductWithIngredientModel) (err error) {
	err = r.db.Where(where).Delete(&ProductWithIngredientModel{}).Error
	return
}

func (r *ProductsWithIngredientsRepo) WriteOffIngredients(productID uint, count int) (err error) {
	//находим связи с ингредиентами, для продукта
	var pwiList []ProductWithIngredientModel
	if err = r.db.Where("product_id = ?", productID).Find(&pwiList).Error; err != nil {
		return err
	}

	//для каждой связи ищем ингредиент. Отнимаем нужное кол-во ингредиента
	for _, pwi := range pwiList {
		var ingredient IngredientModel
		if err = r.db.Where("id = ?", pwi.IngredientID).First(&ingredient).Error; err != nil {
			return err
		}

		ingredient.Count -= pwi.CountTakeForSell * float64(count)
		if err = r.db.Model(&IngredientModel{}).Where(&IngredientModel{Model: gorm.Model{ID: ingredient.ID}}).UpdateColumn("count", ingredient.Count).Error; err != nil {
			return err
		}
	}
	return
}

func (r *ProductsWithIngredientsRepo) ReturnIngredients(productID uint, count int) (err error) {
	var pwiList []ProductWithIngredientModel
	if err = r.db.Where("product_id = ?", productID).Find(&pwiList).Error; err != nil {
		return err
	}

	for _, pwi := range pwiList {
		var ingredient IngredientModel
		if err = r.db.Where("id = ?", pwi.IngredientID).First(&ingredient).Error; err != nil {
			return err
		}

		ingredient.Count += pwi.CountTakeForSell * float64(count)
		if err = r.db.Model(&IngredientModel{}).Where(&IngredientModel{Model: gorm.Model{ID: ingredient.ID}}).UpdateColumn("count", ingredient.Count).Error; err != nil {
			return err
		}
	}
	return
}
