package repository

import "gorm.io/gorm"

//Отчёт о продажах
type ReportRevenueModel struct {
	gorm.Model

	BankEarned float64
	CashEarned float64

	Date int64 // (in unixmilli) за какое число отчёт

	OutletID uint

	OutletModel OutletModel `gorm:"foreignKey:OutletID"`
}

type ReportRevenueRepo struct {
	db *gorm.DB
}

func newReportRevenueRepo(db *gorm.DB) *ReportRevenueRepo {
	return &ReportRevenueRepo{
		db: db,
	}
}

func (r *ReportRevenueRepo) Create(model *ReportRevenueModel) error {
	return r.db.Create(model).Error
}

func (r *ReportRevenueRepo) FindFirts(where *ReportRevenueModel) (result *ReportRevenueModel, err error) {
	err = r.db.Where(where).First(&result).Error
	return
}

func (r *ReportRevenueRepo) Updates(where *ReportRevenueModel, updatedFields *ReportRevenueModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}
