package repository

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OrganizationsRepo struct {
	db *gorm.DB
}

type OrganizationModel struct {
	ID        uint
	CreatedAt time.Time

	Name     string
	Email    string `gorm:"index:,unique"`
	Password string

	EmailConfirmed bool
}

func (r *OrganizationsRepo) generatePasswordHash(pwd string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pwd), 7)
}

func newOrganizationsRepo(db *gorm.DB) *OrganizationsRepo {
	return &OrganizationsRepo{
		db: db,
	}
}

func (r *OrganizationsRepo) Create(m *OrganizationModel) error {
	pwd, err := r.generatePasswordHash(m.Password)
	if err != nil {
		return err
	}
	m.Password = string(pwd)

	return r.db.Create(m).Error
}

func (r *OrganizationsRepo) FindFirts(where *OrganizationModel) (result *OrganizationModel, err error) {
	err = r.db.Where(where).First(&result).Error
	return
}

func (r *OrganizationsRepo) SetPassword(orgID interface{}, password string) error {
	pwd, err := r.generatePasswordHash(password)
	if err != nil {
		return err
	}
	return r.db.Model(&OrganizationModel{}).Where("id = ?", orgID).UpdateColumn("password", string(pwd)).Error
}

func (r *OrganizationsRepo) SetConfirmEmail(email string, val bool) error {
	return r.db.Model(&OrganizationModel{}).Where("email = ?", email).Update("email_confirmed", val).Error
}

func (r *OrganizationsRepo) EmailExists(email string) bool {
	return r.db.Select("id").Where(&OrganizationModel{Email: email}).First(&OrganizationModel{}).Error == nil
}

func (r *OrganizationsRepo) SignIn(email string, password string) (org *OrganizationModel, err error) {
	err = r.db.Where("email = ?", email).First(&org).Error
	if err != nil {
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(org.Password), []byte(password))
	return
}
