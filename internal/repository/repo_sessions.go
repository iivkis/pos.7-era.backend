package repository

import (
	"errors"
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

type SessionsRepo struct {
	db *gorm.DB
}

func newSessionsRepo(db *gorm.DB) *SessionsRepo {
	return &SessionsRepo{
		db: db,
	}
}

//Open - открывает новую сессию, если предыдущие сессии закрыты
func (r *SessionsRepo) Open(m *SessionModel) (err error) {
	//если найдена запись с открытой сессией, то возвращаем ошибку о том, что одновременно можно открыть только одну сессию
	//иначе создаем новую сессию
	if err = r.db.Where("employee_id = ? AND date_close IS NULL", m.EmployeeID).First(&SessionModel{}).Error; err == nil {
		err = ErrSessionAlreadyOpen
		return
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.Create(m).Error
	}
	return
}

func (r *SessionsRepo) GetAllUnscopedByOrgID(orgID uint) (models []SessionModel, err error) {
	err = r.db.Unscoped().Where("org_id = ?", orgID).Order("id desc").Find(&models).Error
	return
}

func (r *SessionsRepo) GetByEmployeeID(employeeID uint) (model SessionModel, err error) {
	err = r.db.Where("employee_id = ?", employeeID).First(&model).Error
	return
}

func (r *SessionsRepo) GetLastForOutlet(outletID uint) (model SessionModel, err error) {
	err = r.db.Unscoped().Where("outlet_id = ? AND date_close IS NOT NULL", outletID).Last(&model).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (r *SessionsRepo) CloseByEmployeeID(employeeID uint, dateClose time.Time, cashClose float64) (err error) {
	if err = r.db.Model(&SessionModel{}).Where("employee_id = ?", employeeID).
		Update("cash_session_close", cashClose).
		Update("date_close", dateClose.String()).Error; err != nil {
		return
	}
	return
}
