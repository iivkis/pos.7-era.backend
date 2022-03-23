package repository

import (
	"gorm.io/gorm"
)

type CashChangesModel struct {
	ID uint

	Date    int64   //unixmilli
	Total   float64 //сумма операции
	Reason  string  //причина сняти\внесения
	Comment string  //комментарий к операции

	SessionID  uint `gorm:"default:NULL"`
	EmployeeID uint
	OutletID   uint
	OrgID      uint

	SessionModel      SessionModel      `gorm:"foreignKey:SessionID"`
	EmployeeModel     EmployeeModel     `gorm:"foreignKey:EmployeeID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type CashChangesRepo struct {
	db *gorm.DB
}

func newCashChangesRepo(db *gorm.DB) *CashChangesRepo {
	return &CashChangesRepo{
		db: db,
	}
}

//actual
func (r *CashChangesRepo) Create(model *CashChangesModel) error {
	return r.db.Create(model).Error
}

func (r *CashChangesRepo) Find(where *CashChangesModel) (result *[]CashChangesModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

func (r CashChangesRepo) FindWithPeriod(dateStart uint64, dateEnd uint64, where *CashChangesModel) (result *[]CashChangesModel, err error) {
	if dateEnd <= 0 {
		err = r.db.Where("date >= ?", dateStart).Find(&result, where).Error
		return
	}
	err = r.db.Where("date >= ? AND date <= ?", dateStart, dateEnd).Find(&result, where).Error
	return
}
