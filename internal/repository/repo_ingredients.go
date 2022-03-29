package repository

import "gorm.io/gorm"

type IngredientModel struct {
	ID uint

	Name          string
	Count         float64
	PurchasePrice float64 //закупочная цена
	MeasureUnit   int     // единица измерения [1 - кг, 2 - л, 3 - шт]

	OutletID uint
	OrgID    uint

	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type IngredientsRepo struct {
	db *gorm.DB
}

func newIngredientsRepo(db *gorm.DB) *IngredientsRepo {
	return &IngredientsRepo{
		db: db,
	}
}

//actual
func (r *IngredientsRepo) Create(ingredient *IngredientModel) error {
	return r.db.Create(ingredient).Error
}

func (r IngredientsRepo) Find(where *IngredientModel) (result *[]IngredientModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r IngredientsRepo) FindFirts(where *IngredientModel) (result *IngredientModel, err error) {
	err = r.db.Where(where).First(&result).Error
	return
}

func (r *IngredientsRepo) Updates(where *IngredientModel, updatedFields *IngredientModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *IngredientsRepo) Delete(where *IngredientModel) (err error) {
	err = r.db.Where(where).Delete(&IngredientModel{}).Error
	return
}

func (r *IngredientsRepo) Exists(where *IngredientModel) bool {
	return r.db.Select("id").Where(where).First(&IngredientModel{}).Error == nil
}
