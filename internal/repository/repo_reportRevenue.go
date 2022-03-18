package repository

import "gorm.io/gorm"

type ReportRevenueModel struct {
	gorm.Model

	BankEarned float64
	CashEarned float64

	Date int64 //in unixmilli
}
