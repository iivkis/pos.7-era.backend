package controller

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func orgSignUp(t *testing.T, engine *gin.Engine) gin.H {
	body := gin.H{
		"name":     "Test",
		"email":    testutil.RandomString(10) + "@test.test",
		"password": testutil.RandomString(10),
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/signUp.Org", testutil.Marshal(body))

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	return body
}

func orgSignIn(t *testing.T, engine *gin.Engine, body gin.H) (data authSignInOrgResponse) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/signIn.Org", testutil.Marshal(body))

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	require.NotEmpty(t, data.Token)
	return
}

func orgGetToken(t *testing.T, engine *gin.Engine) string {
	account := orgSignUp(t, engine)

	body := gin.H{
		"email":    account["email"],
		"password": account["password"],
	}

	data := orgSignIn(t, engine, body)
	require.NotEmpty(t, data.Token)

	return data.Token
}

func TestSignUpOrg(t *testing.T) {
	engine := newController(t)

	var (
		wg sync.WaitGroup
		n  = 11
	)

	wg.Add(n)
	defer wg.Wait()

	for i := 0; i < n; i++ {
		go func() {
			orgSignUp(t, engine)
			wg.Done()
		}()
	}
}

func TestSignInOrg(t *testing.T) {
	engine := newController(t)
	account := orgSignUp(t, engine)

	body := gin.H{
		"email":    account["email"],
		"password": account["password"],
	}

	var (
		wg sync.WaitGroup
		n  = 10
	)

	wg.Add(n)
	defer wg.Wait()

	for i := 0; i < n; i++ {
		go func() {
			orgSignIn(t, engine, body)
			wg.Done()
		}()
	}
}
