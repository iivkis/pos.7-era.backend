package repository

import (
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Repository struct {
	Organizations OrganizationsRepository
}

func NewRepository(authjwt authjwt.AuthJWT) Repository {
	db, err := gorm.Open(sqlite.Open("pos-ninja.sqlite3"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&OrganizationModel{},
	)

	return Repository{
		Organizations: newOrganizationsRepository(db, authjwt),
	}
}
