package repository

import (
	"time"

	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OrganizationsRepository interface {
	Create(m *OrganizationModel) error
	SignIn(email string, password string) (token string, err error)
}

type organizations struct {
	db      *gorm.DB
	authjwt authjwt.AuthJWT
}

type OrganizationModel struct {
	ID        uint
	CreatedAt time.Time
	Name      string
	Email     string `gorm:"unique"`
	Password  string

	EmailConfirmed bool
}

func newOrganizationsRepository(db *gorm.DB, authjwt authjwt.AuthJWT) *organizations {
	return &organizations{
		db:      db,
		authjwt: authjwt,
	}
}

func (r *organizations) Create(m *OrganizationModel) error {

	h, err := bcrypt.GenerateFromPassword([]byte(m.Password), 7)
	if err != nil {
		return err
	}
	m.Password = string(h)
	return r.db.Create(m).Error
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
