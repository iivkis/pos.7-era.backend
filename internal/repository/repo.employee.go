package repository

import (
	"time"

	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"gorm.io/gorm"
)

type EmployeesRepository interface {
	Create(m *EmployeeModel) error
	SignIn(email string, password string, orgID uint) (token string, err error)
	SetPassword(id uint, pwd string) error
}

type employees struct {
	db      *gorm.DB
	authjwt authjwt.AuthJWT
}

func newEmployeeRepo(db *gorm.DB, authjwt authjwt.AuthJWT) *employees {
	return &employees{
		db:      db,
		authjwt: authjwt,
	}
}

type EmployeeModel struct {
	ID        uint
	CreatedAt time.Time

	Name       string
	Surname    string
	Patronymic string

	Email    string `gorm:"unique"`
	Password string

	OrgID  uint
	RoleID int

	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (r *employees) Create(m *EmployeeModel) error {
	if err := r.db.Create(m).Error; err != nil {
		return err
	}

	if err := r.SetPassword(m.ID, m.Password); err != nil {
		return err
	}

	return nil
}

func (r *employees) SignIn(email string, password string, orgID uint) (token string, err error) {
	var model EmployeeModel
	if err = r.db.Where("email = ? AND org_id = ? AND password = ?", email, orgID, password).First(&model).Error; err != nil {
		return "", err
	}

	claims := authjwt.EmployeeClaims{
		OrganizationID: orgID,
		EmployeeID:     model.ID,
	}

	token, err = r.authjwt.SignInEmployee(&claims)
	if err != nil {
		return "", err
	}

	return token, err
}

func (r *employees) SetPassword(id uint, pwd string) error {
	if err := r.db.Model(&EmployeeModel{}).Where("id = ?", id).Update("password", pwd).Error; err != nil {
		return err
	}
	return nil
}
