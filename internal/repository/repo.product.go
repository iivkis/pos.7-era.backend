package repository

import "gorm.io/gorm"

type ProductModel struct {
	gorm.Model

	Name       string
	CategoryID uint

	Amount int
	Price  float64

	Photo string

	CategoryModel CategoryModel `gorm:"foreignKey:CategoryID"`
}
