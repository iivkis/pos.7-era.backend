package authjwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AuthJWT interface {
	SignInOrganization(claims *OrganizationClaims) (token string, err error)
}

type authjwt struct {
	secret []byte
}

type OrganizationClaims struct {
	OrganizationID uint  `json:"organization_id"`
	CreatedAt      int64 `json:"created_at"`
	jwt.StandardClaims
}

func NewAuthJWT(secret []byte) *authjwt {
	return &authjwt{
		secret: secret,
	}
}

func (t *authjwt) SignInOrganization(claims *OrganizationClaims) (token string, err error) {
	claims.Issuer = "pos-ninja.ru"
	claims.CreatedAt = time.Now().UTC().Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(t.secret)
}
