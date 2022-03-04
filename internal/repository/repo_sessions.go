package repository

import (
	"errors"

	"gorm.io/gorm"
)

type SessionModel struct {
	ID uint

	DateOpen  int64
	DateClose int64

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
	return err
}

//Возвращает все сессии организации (в том числе удаленные)
func (r *SessionsRepo) GetAllByOrgID(orgID uint) (models []SessionModel, err error) {
	err = r.db.Unscoped().Where("org_id = ?", orgID).Order("id desc").Find(&models).Error
	return
}

//Возвращает все сессии организации (в том числе удаленные)
func (r *SessionsRepo) GetAllByOutletID(outletID uint) (models []SessionModel, err error) {
	err = r.db.Unscoped().Where("outlet_id = ?", outletID).Order("id desc").Find(&models).Error
	return
}

//Возвращает последнюю сессию сотрудника
func (r *SessionsRepo) GetLastOpenByEmployeeID(employeeID interface{}) (model SessionModel, err error) {
	err = r.db.Where("employee_id = ? AND date_close = 0", employeeID).Last(&model).Error
	return
}

//Возвращает последнюю закрытую сессию для точки продаж
func (r *SessionsRepo) GetLastClosedForOutlet(outletID uint) (model SessionModel, err error) {
	err = r.db.Where("outlet_id = ? AND date_close <> 0", outletID).Last(&model).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

//Возвращает последнюю сессию для точки продаж
func (r *SessionsRepo) GetLastForOutlet(outletID uint) (model SessionModel, err error) {
	err = r.db.Where("outlet_id = ?", outletID).Last(&model).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

//при закрытие сессии обновляем поля
// cash_session_close и date_close
func (r *SessionsRepo) Close(employeeID interface{}, sess *SessionModel) error {
	return r.db.Model(&SessionModel{}).Where("employee_id = ? AND date_close <> 0", employeeID).
		Updates(sess).
		Last(sess).Error
}

func (r *SessionsRepo) HasOpenSession(employeeID interface{}) (ok bool, err error) {
	err = r.db.Where("employee_id = ? AND date_close = 0", employeeID).First(&SessionModel{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *SessionsRepo) ExistsWithEmployeeID(sessionID interface{}, employeeID interface{}) bool {
	return r.db.Where("id = ? AND employee_id = ?", sessionID, employeeID).First(&SessionModel{}).Error == nil
}
