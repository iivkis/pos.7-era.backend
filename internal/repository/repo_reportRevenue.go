package repository

import "gorm.io/gorm"

//Отчёт о продажах
type ReportRevenueModel struct {
	gorm.Model

	BankEarned  float64 //заработано в вирт. валюте
	CashEarned  float64 //заработано наличными
	TotalAmount float64 // общая выручка

	NumberOfReceipts int     //кол-во чеков
	AverageReceipt   float64 //средняя сумма чека

	Date int64 // (in unixmilli) за какое число отчёт

	OutletID uint
	OrgID    uint

	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
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

func (r *ReportRevenueRepo) Find(where *ReportRevenueModel) (result *[]ReportRevenueModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r *ReportRevenueRepo) FindWithPeriod(where *InventoryHistoryModel, start uint64, end uint64) (result *[]ReportRevenueModel, err error) {
	if end == 0 {
		err = r.db.Where("date >= ?", start).Find(&result, where).Error
	} else {
		err = r.db.Where("date >= ? AND date <= ?", start, end).Find(&result, where).Error
	}
	return
}

func (r *ReportRevenueRepo) FindFirts(where *ReportRevenueModel) (result *ReportRevenueModel, err error) {
	err = r.db.Where(where).First(&result).Error
	return
}

func (r *ReportRevenueRepo) Updates(where *ReportRevenueModel, updatedFields *ReportRevenueModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}
