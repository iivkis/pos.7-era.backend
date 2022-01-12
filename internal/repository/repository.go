package repository

import (
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Repository struct {
	Organizations OrganizationsRepository
	Employees     EmployeesRepository
}

func NewRepository(authjwt authjwt.AuthJWT) Repository {
	url := "f0602327_posninja:AjKZwdoH@tcp(141.8.193.236)/f0602327_posninja?parseTime=True"
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&OrganizationModel{},
		&EmployeeModel{},
	)

	return Repository{
		Organizations: newOrganizationsRepo(db, authjwt),
		Employees:     newEmployeeRepo(db, authjwt),
	}
}
