package repository

import (
	"errors"

	"gorm.io/gorm"
)

type SessionModel struct {
	gorm.Model

	DateOpen  int64
	DateClose int64

	CashSessionOpen  float64
	CashSessionClose float64

	CashEarned float64
	BankEarned float64

	AddedToReport bool //добавлена ли сессия в отчет

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

//actual

//при закрытие сессии обновляем поля
// cash_session_close и date_close
func (r *SessionsRepo) Close(employeeID interface{}, sess *SessionModel) error {
	if err := r.db.Model(&SessionModel{}).Where("employee_id = ? AND date_close = 0", employeeID).Updates(sess).Error; err != nil {
		return err
	}
	return r.db.Model(&SessionModel{}).Where("employee_id = ?", employeeID).Last(sess).Error
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

func (r *SessionsRepo) HasOpenSession(employeeID interface{}) (ok bool, err error) {
	err = r.db.Where("employee_id = ? AND date_close = 0", employeeID).First(&SessionModel{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *SessionsRepo) Find(where *SessionModel) (result *[]SessionModel, err error) {
	err = r.db.Where(where).Find(&result).Error
	return
}

//поиск outlet_id, где есть новые сессии
func (r *SessionsRepo) FindOutletIDForReport(where *SessionModel) (result *[]uint, err error) {
	err = r.db.Table("session_models").Select("outlet_id").
		Where("(added_to_report = 0 OR added_to_report IS NULL) AND date_close <> 0").
		Distinct("outlet_id").Find(&result, where).Error
	return
}

//поиск новых сесстий для отчёта
func (r *SessionsRepo) FindSessionsForReport(where *SessionModel) (result *[]SessionModel, err error) {
	err = r.db.Where("(added_to_report = 0 OR added_to_report IS NULL) AND date_close <> 0").Order("date_close").Find(&result, where).Error
	return
}

func (r *SessionsRepo) SetFieldAddedToReport(val bool, sessionID []uint) (err error) {
	err = r.db.Model(&SessionModel{}).Where("id IN ?", sessionID).UpdateColumn("added_to_report", val).Error
	return
}

func (r *SessionsRepo) Exists(where *SessionModel) bool {
	return r.db.Where(where).First(&SessionModel{}).Error == nil
}
