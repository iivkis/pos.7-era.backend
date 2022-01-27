package repository

import "gorm.io/gorm"

type CategoryModel struct {
	gorm.Model
	Name     string
	OutletID uint

	OutletModel OutletModel `gorm:"foreignKey:OutletID"`
}
