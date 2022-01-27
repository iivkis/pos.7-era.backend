package repository

import (
	"time"

	"gorm.io/gorm"
)

type SessionModel struct {
	ID uint

	DateOpen  time.Time
	DateClose gorm.DeletedAt

	CashSessionOpen  float64
	CashSessionClose float64

	EmployeeID uint `gorm:"index"`
	OutletID   uint `gorm:"index"`

	EmployeeModel EmployeeModel `gorm:"foreignKey:EmployeeID"`
	OutletModel   OutletModel   `gorm:"foreignKey:OutletID"`
}

type SessionsRepository interface {
	Open(m *SessionModel) error
	CloseByEmployeeID(employeeID uint, cashClose float64) (err error)
	GetAllUnscoped() (models []SessionModel, err error)
	GetLastForOutlet(outletID uint) (model SessionModel, err error)
}

type sessions struct {
	db *gorm.DB
}

func newSessionsRepo(db *gorm.DB) *sessions {
	return &sessions{
		db: db,
	}
}

func (r *sessions) Open(m *SessionModel) error {
	m.DateOpen = time.Now().UTC()
	return r.db.Create(m).Error
}

func (r *sessions) GetAllUnscoped() (models []SessionModel, err error) {
	err = r.db.Unscoped().Order("id desc").Find(&models).Error
	return
}

func (r *sessions) GetByEmployeeID(employeeID uint) (model SessionModel, err error) {
	err = r.db.Where("employee_id = ?", employeeID).First(&model).Error
	return
}

func (r *sessions) GetLastForOutlet(outletID uint) (model SessionModel, err error) {
	err = r.db.Unscoped().Where("outlet_id = ? AND date_close IS NOT NULL", outletID).Last(&model).Error
	return
}

func (r *sessions) CloseByEmployeeID(employeeID uint, cashClose float64) (err error) {
	if err = r.db.Model(&SessionModel{}).Where("employee_id = ?", employeeID).Update("cash_session_close", cashClose).Error; err != nil {
		return
	}
	err = r.db.Where("employee_id = ?", employeeID).Delete(&SessionModel{}).Error
	return
}
