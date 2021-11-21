package repository

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Repository struct {
	Organizations iOrganizations
}

func NewRepository() Repository {
	db, err := gorm.Open(sqlite.Open("pos-ninja.sqlite3"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&OrganizationModel{},
	)

	return Repository{
		Organizations: newOrganizationsRepo(db),
	}
}
