package authjwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var errInvalidToken = errors.New("invalid token")

type AuthJWT interface {
	SignInOrganization(claims *OrganizationClaims) (token string, err error)
	SignInEmployee(claims *EmployeeClaims) (token string, err error)

	ParseOrganizationToken(token string) (*OrganizationClaims, error)
	ParseEmployeeToken(token string) (*EmployeeClaims, error)
}

type authjwt struct {
	secretOrg      []byte
	secretEmployee []byte
}

type OrganizationClaims struct {
	OrganizationID uint  `json:"organization_id"`
	CreatedAt      int64 `json:"created_at"`
	jwt.StandardClaims
}

type EmployeeClaims struct {
	OrganizationID uint   `json:"organization_id"`
	EmployeeID     uint   `json:"employee_id"`
	OutletID       uint   `json:"outlet_id"`
	Role           string `json:"role"`
	CreatedAt      int64  `json:"created_at"`
	jwt.StandardClaims
}

func NewAuthJWT(secret []byte) *authjwt {
	return &authjwt{
		secretOrg:      secret,
		secretEmployee: reverse(secret),
	}
}

func reverse(arr []byte) (rev []byte) {
	rev = make([]byte, len(arr))
	for i := range arr {
		rev[len(arr)-1-i] = arr[i]
	}
	return
}

func (j *authjwt) SignInOrganization(claims *OrganizationClaims) (token string, err error) {
	claims.Issuer = "pos-ninja.ru"
	claims.CreatedAt = time.Now().UTC().Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(j.secretOrg)
}

func (j *authjwt) SignInEmployee(claims *EmployeeClaims) (token string, err error) {
	claims.Issuer = "pos-ninja.ru"
	claims.CreatedAt = time.Now().UTC().Unix()
	claims.ExpiresAt = time.Now().Unix() + 60*60*24 //токен живет 24 часа
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(j.secretEmployee)
}

func (j *authjwt) ParseOrganizationToken(token string) (*OrganizationClaims, error) {
	t, err := jwt.ParseWithClaims(token, &OrganizationClaims{}, func(t *jwt.Token) (interface{}, error) {
		return j.secretOrg, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(*OrganizationClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, errInvalidToken
}

func (j *authjwt) ParseEmployeeToken(token string) (*EmployeeClaims, error) {
	t, err := jwt.ParseWithClaims(token, &EmployeeClaims{}, func(t *jwt.Token) (interface{}, error) {
		return j.secretEmployee, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(*EmployeeClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, errInvalidToken
}
