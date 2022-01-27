package repository

import "gorm.io/gorm"

type OrderInfoModel struct {
	gorm.Model
	PayType      int // 1 - наличные, 2 - безналичные, 3 - смешанный
	EmployeeName string
	SessionID    uint

	SessionModel SessionModel `gorm:"foreignKey:SessionID"`
}
