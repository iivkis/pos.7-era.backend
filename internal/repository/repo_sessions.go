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
//если найдена запись с открытой сессией, то возвращаем ошибку о том, что одновременно можно открыть только одну сессию
//иначе создаем новую сессию
func (r *SessionsRepo) Open(m *SessionModel) error {
	// if err = r.db.Where("employee_id = ? AND date_close IS NULL", m.EmployeeID).First(&SessionModel{}).Error; err == nil {
	// 	err = ErrSessionAlreadyOpen
	// 	return
	// } else if errors.Is(err, gorm.ErrRecordNotFound) {
	// 	err = r.db.Create(m).Error
	// }
	ok, err := r.HasOpenSession(m.EmployeeID)
	if err != nil {
		return err
	}

	if ok {
		return ErrSessionAlreadyOpen
	}

	err = r.db.Create(m).Error
	if err != nil {
		return err
	}

	err = r.db.Model(&EmployeeModel{}).Where("id = ?", m.EmployeeID).Update("online", true).Error
	return err
}

//Возвращает все сессии организации (в том числе удаленные)
func (r *SessionsRepo) GetAllUnscopedByOrgID(orgID uint) (models []SessionModel, err error) {
	err = r.db.Unscoped().Where("org_id = ?", orgID).Order("id desc").Find(&models).Error
	return
}

//Возвращает сессию сотрудника
func (r *SessionsRepo) GetByEmployeeID(employeeID uint) (model SessionModel, err error) {
	err = r.db.Where("employee_id = ?", employeeID).First(&model).Error
	return
}

//Возвращает последнюю закрытую сессию для точки продаж
func (r *SessionsRepo) GetLastClosedForOutlet(outletID uint) (model SessionModel, err error) {
	err = r.db.Unscoped().Where("outlet_id = ? AND date_close IS NOT NULL", outletID).Last(&model).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

//закрывает сессию сотрудника
func (r *SessionsRepo) CloseByEmployeeID(employeeID uint, dateClose time.Time, cashClose float64) (err error) {
	if err = r.db.Model(&SessionModel{}).Where("employee_id = ?", employeeID).
		Update("cash_session_close", cashClose).
		Update("date_close", dateClose.String()).Error; err != nil {
		return
	}

	err = r.db.Model(&EmployeeModel{}).Where("id = ?", employeeID).Update("online", false).Error
	return
}

func (r *SessionsRepo) HasOpenSession(employeeID interface{}) (ok bool, err error) {
	err = r.db.Where("employee_id = ? AND date_close IS NULL", employeeID).First(&SessionModel{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
