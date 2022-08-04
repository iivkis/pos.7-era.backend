package repository

import (
	"database/sql"

	"gorm.io/gorm"
)

type ProductWithIngredientModel struct {
	DeletedAt gorm.DeletedAt

	ID uint

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

// actual
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

func (r *ProductsWithIngredientsRepo) SubractionIngredients(productID uint, count int) (err error) {
	//находим связи с ингредиентами, для продукта
	var pwiList []ProductWithIngredientModel
	if err = r.db.Where(&ProductWithIngredientModel{ProductID: productID}).Find(&pwiList).Error; err != nil {
		return err
	}

	//для каждой связи ищем ингредиент. Отнимаем нужное кол-во ингредиента
	for _, pwi := range pwiList {
		if err := r.db.Exec("UPDATE `ingredient_models` SET `count` = `count` - @n WHERE `id` = @id",
			sql.Named("n", pwi.CountTakeForSell*float64(count)),
			sql.Named("id", pwi.IngredientID),
		).Error; err != nil {
			return err
		}
	}
	return
}

func (r *ProductsWithIngredientsRepo) AdditionIngredients(productID uint, count int) (err error) {
	//находим связи с ингредиентами, для продукта
	var pwiList []ProductWithIngredientModel
	if err = r.db.Where(&ProductWithIngredientModel{ProductID: productID}).Find(&pwiList).Error; err != nil {
		return err
	}

	//для каждой связи ищем ингредиент. Прибавляем нужное кол-во ингредиента
	for _, pwi := range pwiList {
		if err := r.db.Exec("UPDATE `ingredient_models` SET `count` = `count` + @n WHERE `id` = @id",
			sql.Named("n", pwi.CountTakeForSell*float64(count)),
			sql.Named("id", pwi.IngredientID),
		).Error; err != nil {
			return err
		}
	}
	return
}
