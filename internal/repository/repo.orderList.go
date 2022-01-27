package repository

import "gorm.io/gorm"

type OrderListModel struct {
	gorm.Model
	ProductName string

	ProductPrice float32
	Count        int

	ProductID uint
	OrderID   uint

	ProductModel   ProductModel   `gorm:"foreignKey:ProductID"`
	OrderInfoModel OrderInfoModel `gorm:"foreignKey:OrderID"`
}
