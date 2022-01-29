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
	OrgID      uint `gorm:"index"`

	EmployeeModel     EmployeeModel     `gorm:"foreignKey:EmployeeID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
}

type SessionsRepository interface {
	Open(m *SessionModel) error
	CloseByEmployeeID(employeeID uint, dateClose time.Time, cashClose float64) (err error)
	GetAllUnscopedByOrgID(orgID uint) (models []SessionModel, err error)
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
	return r.db.Create(m).Error
}

func (r *sessions) GetAllUnscopedByOrgID(orgID uint) (models []SessionModel, err error) {
	err = r.db.Unscoped().Where("org_id = ?", orgID).Order("id desc").Find(&models).Error
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

func (r *sessions) CloseByEmployeeID(employeeID uint, dateClose time.Time, cashClose float64) (err error) {
	if err = r.db.Model(&SessionModel{}).Where("employee_id = ?", employeeID).
		Update("cash_session_close", cashClose).
		Update("date_close", dateClose.String()).Error; err != nil {
		return
	}
	return
}