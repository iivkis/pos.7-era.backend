package repository

import (
	"fmt"

	"github.com/iivkis/pos-ninja-backend/internal/config"
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Repository struct {
	Organizations *OrganizationsRepo
	Employees     *EmployeesRepo
	Outlets       *OutletsRepo
	Sessions      *SessionsRepo
	Category      *CategoriesRepo
}

func NewRepository(authjwt *authjwt.AuthJWT) Repository {
	url := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True", config.Env.DatabaseLogin, config.Env.DatabasePassword, config.Env.DatabaseIP, config.Env.DatabaseLogin)
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&OrganizationModel{},
		&EmployeeModel{},
		&OutletModel{},
		&SessionModel{},
		&ProductModel{},
		&OrderInfoModel{},
		&OrderListModel{},
		&CategoryModel{},
	)

	return Repository{
		Organizations: newOrganizationsRepo(db),
		Employees:     newEmployeesRepo(db),
		Outlets:       newOutletsRepo(db),
		Sessions:      newSessionsRepo(db),
		Category:      newCategoriesRepo(db),
	}
}
