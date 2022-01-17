package repository

import (
	"errors"
	"time"

	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OrganizationsRepository interface {
	Create(m *OrganizationModel) error
	SignIn(email string, password string) (token string, err error)
	ConfirmEmailTrue(email string) error
	EmailExists(email string) (ok bool, err error)
	SetPassword(id uint, pwd string) error
}

type organizations struct {
	db      *gorm.DB
	authjwt authjwt.AuthJWT
}

type OrganizationModel struct {
	ID        uint
	CreatedAt time.Time

	Name     string
	Email    string `gorm:"unique"`
	Password string

	EmailConfirmed bool
}

func newOrganizationsRepo(db *gorm.DB, authjwt authjwt.AuthJWT) *organizations {
	return &organizations{
		db:      db,
		authjwt: authjwt,
	}
}

func (r *organizations) Create(m *OrganizationModel) error {
	if err := r.db.Create(m).Error; err != nil {
		return err
	}

	if err := r.SetPassword(m.ID, m.Password); err != nil {
		return err
	}

	return nil
}

func (r *organizations) SetPassword(id uint, pwd string) error {
	cpwd, err := bcrypt.GenerateFromPassword([]byte(pwd), 7)
	if err != nil {
		return err
	}

	if err := r.db.Model(&OrganizationModel{}).Where("id = ?", id).Update("password", string(cpwd)).Error; err != nil {
		return err
	}
	return nil
}

func (r *organizations) ConfirmEmailTrue(email string) error {
	return r.db.Model(&OrganizationModel{}).Where("email = ?", email).Update("email_confirmed", true).Error
}

func (r *organizations) EmailExists(email string) (bool, error) {
	err := r.db.Model(&OrganizationModel{}).Where("email = ?", email).First(nil).Error
	if err == nil {
		return true, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (r *organizations) SignIn(email string, password string) (token string, err error) {
	var model OrganizationModel
	if err = r.db.Where("email = ?", email).First(&model).Error; err != nil {
		return "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(model.Password), []byte(password)); err != nil {
		return "", err
	}

	claims := authjwt.OrganizationClaims{
		OrganizationID: model.ID,
	}

	token, err = r.authjwt.SignInOrganization(&claims)
	if err != nil {
		return "", err
	}

	return token, err
}
