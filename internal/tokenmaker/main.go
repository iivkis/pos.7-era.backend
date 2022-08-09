package tokenmaker

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var errInvalidToken = errors.New("invalid token")

type TokenMaker struct {
	organizationSecretKey []byte
	employeeSecretKey     []byte
}

type OrganizationClaims struct {
	OrganizationID uint  `json:"organization_id"`
	CreatedAt      int64 `json:"created_at"`

	jwt.StandardClaims
}

type EmployeeClaims struct {
	EmployeeID     uint `json:"employee_id"`
	OutletID       uint `json:"outlet_id"`
	OrganizationID uint `json:"organization_id"`

	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`

	jwt.StandardClaims
}

// проверка, имеет ли сотрудник какую-либо роль из массива roles
func (m *EmployeeClaims) HasRole(roles ...string) bool {
	for _, role := range roles {
		if role == m.Role {
			return true
		}
	}
	return false
}

func NewTokenMaker(secretKey []byte) *TokenMaker {
	return &TokenMaker{
		organizationSecretKey: secretKey,
		employeeSecretKey:     reverse(secretKey),
	}
}

func reverse(arr []byte) (rev []byte) {
	rev = make([]byte, len(arr))
	for i := range arr {
		rev[len(arr)-1-i] = arr[i]
	}
	return
}

func (j *TokenMaker) CreateOrganizationToken(claims *OrganizationClaims) (token string, err error) {
	claims.Issuer = "pos-7era"
	claims.CreatedAt = time.Now().UTC().Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(j.organizationSecretKey)
}

func (j *TokenMaker) CreateEmployeeToken(claims *EmployeeClaims) (token string, err error) {
	claims.Issuer = "pos-7era"
	claims.CreatedAt = time.Now().UTC().Unix()
	claims.ExpiresAt = time.Now().Add(time.Hour * 24).Unix() //токен живет 24 часа
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(j.employeeSecretKey)
}

func (j *TokenMaker) ParseOrganizationToken(token string) (*OrganizationClaims, error) {
	t, err := jwt.ParseWithClaims(token, &OrganizationClaims{}, func(t *jwt.Token) (interface{}, error) {
		return j.organizationSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(*OrganizationClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, errInvalidToken
}

func (j *TokenMaker) ParseEmployeeToken(token string) (*EmployeeClaims, error) {
	t, err := jwt.ParseWithClaims(token, &EmployeeClaims{}, func(t *jwt.Token) (interface{}, error) {
		return j.employeeSecretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(*EmployeeClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, errInvalidToken
}
