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
	SignIn(email string, password string) (org OrganizationModel, err error)
	ConfirmEmailTrue(email string) error
	EmailExists(email string) (ok bool, err error)
	SetPassword(id uint, pwd string) error
}

type organizations struct {
	db      *gorm.DB
	authjwt *authjwt.AuthJWT
}

type OrganizationModel struct {
	ID        uint
	CreatedAt time.Time

	Name     string
	Email    string `gorm:"index:,unique"`
	Password string

	EmailConfirmed bool
}

func newOrganizationsRepo(db *gorm.DB, authjwt *authjwt.AuthJWT) *organizations {
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
	err := r.db.Where("email = ?", email).First(&OrganizationModel{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *organizations) SignIn(email string, password string) (org OrganizationModel, err error) {
	if err = r.db.Where("email = ?", email).First(&org).Error; err != nil {
		return OrganizationModel{}, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(org.Password), []byte(password)); err != nil {
		return OrganizationModel{}, err
	}
	return
}
