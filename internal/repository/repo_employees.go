package repository

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

type EmployeeModel struct {
	ID        uint
	CreatedAt time.Time

	Name     string
	Password string

	OrgID    uint
	OutletID uint
	Role     string

	DeletedAt gorm.DeletedAt `gorm:"index"`

	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
}

type EmployeesRepo struct {
	db *gorm.DB
}

func newEmployeesRepo(db *gorm.DB) *EmployeesRepo {
	return &EmployeesRepo{
		db: db,
	}
}

func (r *EmployeesRepo) Create(m *EmployeeModel) error {
	if !roleIsExists(m.Role) {
		return ErrUndefinedRole
	}

	if err := r.checkPasswordCorret(m.Password); err != nil {
		return ErrOnlyNumInPassword
	}

	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	return nil
}

func (r *EmployeesRepo) SignIn(id uint, password string, orgID uint) (empl EmployeeModel, err error) {
	err = r.db.Where("id = ? AND org_id = ? AND password = ?", id, orgID, password).First(&empl).Error
	return
}

func (r *EmployeesRepo) SetPassword(id uint, pwd string) (err error) {
	if err = r.checkPasswordCorret(pwd); err != nil {
		return
	}
	err = r.db.Model(&EmployeeModel{}).Where("id = ?", id).Update("password", pwd).Error
	return
}

func (r *EmployeesRepo) GetAll(orgID uint) ([]EmployeeModel, error) {
	var models []EmployeeModel
	if err := r.db.Where("org_id = ?", orgID).Find(&models).Error; err != nil {
		return models, err
	}
	return models, nil
}

func (r *EmployeesRepo) checkPasswordCorret(pwd string) error {
	n, err := strconv.Atoi(pwd)
	if err != nil || n < 0 {
		return ErrOnlyNumInPassword
	}
	return nil
}
