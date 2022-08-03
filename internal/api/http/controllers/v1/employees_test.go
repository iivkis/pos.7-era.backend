package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func employessGetAll(t *testing.T, engine *gin.Engine, token string) (data employeesGetAllResponse) {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/api/v1/employees", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	return
}

func employeeGetOwnerToken(t *testing.T, engine *gin.Engine, tokenOrg string) (tokenEmployee string) {
	employees := employessGetAll(t, engine, tokenOrg)

	body := gin.H{
		"id":       employees[0].ID,
		"password": "000000",
	}

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/api/v1/auth/signIn.Employee", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, tokenOrg)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var (
		response Response
		data     authSignInEmployeeResponse
	)

	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	return data.Token
}

func TestEmployeesGetAll(t *testing.T) {
	engine := newController(t)
	tokenOrg := orgGetToken(t, engine)

	employessGetAll(t, engine, tokenOrg)
}

func TestEmployeeSignIn(t *testing.T) {
	engine := newController(t)
	tokenOrg := orgGetToken(t, engine)

	t.Run("owner", func(t *testing.T) {
		employeeGetOwnerToken(t, engine, tokenOrg)
	})
}
