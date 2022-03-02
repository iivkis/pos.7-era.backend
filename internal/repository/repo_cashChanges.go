package repository

import "gorm.io/gorm"

type CashChangesModel struct {
	ID uint

	Date    int64   //unixmilli
	Total   float64 //сумма операции
	Reason  int     //причина снятия, либо внесения; 0 - снятие; 1 - вклад; 2 -
	Comment string  //комментарий к операции

	SessionID  uint
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

func (r *CashChangesRepo) Create(m *CashChangesModel) error {
	return r.db.Create(m).Error
}

func (r *CashChangesRepo) FindAllBySessionID(sessionID interface{}) (m []CashChangesModel, err error) {
	err = r.db.Where("session_id = ?", sessionID).Find(&m).Error
	return
}

func (r CashChangesRepo) FindAllByOutletID(outletID interface{}) (m []CashChangesModel, err error) {
	err = r.db.Where("outlet_id = ?", outletID).Find(&m).Error
	return
}

func (r CashChangesRepo) FindAllByOutletIDWithPeriod(dateStart int64, dateEnd int64, outletID interface{}) (m []CashChangesModel, err error) {
	err = r.db.Where("date >= ? AND date <= ? AND outlet_id = ?", dateStart, dateEnd, outletID).Find(&m).Error
	return
}
