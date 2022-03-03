package repository

import "gorm.io/gorm"

type IngredientModel struct {
	gorm.Model

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

func (r *IngredientsRepo) Create(ingredient *IngredientModel) error {
	return r.db.Create(ingredient).Error
}

func (r *IngredientsRepo) GetAllByOrgID(orgID interface{}) (ingredients []IngredientModel, err error) {
	err = r.db.Where("org_id = ?", orgID).Find(&ingredients).Error
	return
}

func (r *IngredientsRepo) GetAllByOutletID(outletID interface{}) (ingredients []IngredientModel, err error) {
	err = r.db.Where("outlet_id = ?", outletID).Find(&ingredients).Error
	return
}

func (r *IngredientsRepo) Updates(ingredient *IngredientModel, ingredientID interface{}, outletID interface{}) error {
	return r.db.Where("id = ? AND outlet_id = ?", ingredientID, outletID).
		UpdateColumn("name", ingredient.Name).
		UpdateColumn("count", ingredient.Count).
		UpdateColumn("purchase_price", ingredient.PurchasePrice).
		UpdateColumn("measure_unit", ingredient.MeasureUnit).
		Error
}

func (r *IngredientsRepo) Delete(ingredientID interface{}, outletID interface{}) error {
	return r.db.Where("id = ? AND outlet_id = ?", ingredientID, outletID).Delete(&IngredientModel{}).Error
}

func (r *IngredientsRepo) ExistsInOutlet(ingredientID interface{}, outletID interface{}) bool {
	err := r.db.Where("id = ? AND outlet_id = ?", ingredientID, outletID).First(&IngredientModel{}).Error
	return err == nil
}
