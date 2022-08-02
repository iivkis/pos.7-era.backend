package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/stretchr/testify/require"
)

func newOrgAccount(t *testing.T) gin.H {
	body := gin.H{
		"name":     "Test",
		"email":    testutil.RandomString(10) + "@test.test",
		"password": testutil.RandomString(10),
	}

	engine := newController(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/signUp.Org", testutil.Marshal(body))

	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	return body
}

func TestSignUpOrg(t *testing.T) {
	newOrgAccount(t)
}

func TestSignInOrg(t *testing.T) {
	account := newOrgAccount(t)

	body := gin.H{
		"email":    account["email"],
		"password": account["password"],
	}

	engine := newController(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/signIn.Org", testutil.Marshal(body))

	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestSignInEmployee(t *testing.T) {
	t.Run("owner", func(t *testing.T) {

	})
}
