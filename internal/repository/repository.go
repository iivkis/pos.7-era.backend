package repository

import (
	"fmt"
	"log"

	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Repository struct {
	Organizations            *OrganizationsRepo
	Employees                *EmployeesRepo
	Outlets                  *OutletsRepo
	Sessions                 *SessionsRepo
	Categories               *CategoriesRepo
	Products                 *ProductsRepo
	Ingredients              *IngredientsRepo
	OrdersList               *OrderListRepo
	OrdersInfo               *OrderInfoRepo
	ProductsWithIngredients  *ProductsWithIngredientsRepo
	CashChanges              *CashChangesRepo
	InventoryHistory         *InventoryHistoryRepo
	InventoryList            *InventoryListRepo
	IngredientsAddingHistory *IngredientsAddingHistoryRepo
}

func NewRepository(authjwt *authjwt.AuthJWT) *Repository {
	url := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True", config.Env.DatabaseLogin, config.Env.DatabasePassword, config.Env.DatabaseIP, config.Env.DatabaseName)

	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if *config.Flags.Migration {
		if err := db.AutoMigrate(
			&OrganizationModel{},
			&EmployeeModel{},
			&OutletModel{},
			&SessionModel{},
			&ProductModel{},
			&OrderInfoModel{},
			&OrderListModel{},
			&CategoryModel{},
			&IngredientModel{},
			&ProductWithIngredientModel{},
			&CashChangesModel{},
			&InventoryHistoryModel{},
			&InventoryListModel{},
			&IngredientsAddingHistoryModel{},
		); err != nil {
			panic(err)
		}
		log.Println("migration done")
	}

	return &Repository{
		Organizations:            newOrganizationsRepo(db),
		Employees:                newEmployeesRepo(db),
		Outlets:                  newOutletsRepo(db),
		Sessions:                 newSessionsRepo(db),
		Categories:               newCategoriesRepo(db),
		Products:                 newProductsRepo(db),
		Ingredients:              newIngredientsRepo(db),
		OrdersList:               newOrderListRepo(db),
		OrdersInfo:               newOrderInfoRepo(db),
		ProductsWithIngredients:  newProductsWithIngredientsRepo(db),
		CashChanges:              newCashChangesRepo(db),
		InventoryHistory:         newInventoryHistoryRepo(db),
		InventoryList:            newInventoryListRepo(db),
		IngredientsAddingHistory: newIngredientsAddingHistoryRepo(db),
	}
}
