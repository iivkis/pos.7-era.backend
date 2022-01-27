package repository

import (
	"strconv"
	"time"

	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
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

type EmployeesRepository interface {
	Create(m *EmployeeModel) error
	SignIn(id uint, password string, orgID uint) (empl EmployeeModel, err error)
	SetPassword(id uint, pwd string) error
	GetAll(orgID uint) ([]EmployeeModel, error)
}

type employees struct {
	db *gorm.DB
	// authjwt authjwt.AuthJWT
}

func newEmployeesRepo(db *gorm.DB, authjwt authjwt.AuthJWT) *employees {
	return &employees{
		db: db,
		// authjwt: authjwt,
	}
}

func (r *employees) Create(m *EmployeeModel) error {
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

func (r *employees) SignIn(id uint, password string, orgID uint) (empl EmployeeModel, err error) {
	if err = r.db.Where("id = ? AND org_id = ? AND password = ?", id, orgID, password).First(&empl).Error; err != nil {
		return empl, err
	}

	// claims := authjwt.EmployeeClaims{
	// 	OrganizationID: orgID,
	// 	EmployeeID:     empl.ID,
	// 	Role:           empl.Role,
	// }

	// token, err = r.authjwt.SignInEmployee(&claims)
	// if err != nil {
	// 	return "", err
	// }

	return empl, err
}

func (r *employees) SetPassword(id uint, pwd string) error {
	if err := r.checkPasswordCorret(pwd); err != nil {
		return ErrOnlyNumInPassword
	}

	if err := r.db.Model(&EmployeeModel{}).Where("id = ?", id).Update("password", pwd).Error; err != nil {
		return err
	}
	return nil
}

func (r *employees) GetAll(orgID uint) ([]EmployeeModel, error) {
	var models []EmployeeModel
	if err := r.db.Where("org_id = ?", orgID).Find(&models).Error; err != nil {
		return models, err
	}
	return models, nil
}

func (r *employees) checkPasswordCorret(pwd string) error {
	n, err := strconv.Atoi(pwd)
	if err != nil || n < 0 {
		return ErrOnlyNumInPassword
	}
	return nil
}
