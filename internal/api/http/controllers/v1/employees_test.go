package controller

import (
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestEmployeeUpdate(t *testing.T) {
	engine, w := newController(t), httptest.NewRecorder()

	tokenOrg := orgGetToken(t, engine)
	tokenOwner := employeeGetOwnerToken(t, engine, tokenOrg)

	employee := employessGetAll(t, engine, tokenOrg)[1] //автоматически созданный кассир

	id := strconv.Itoa(int(employee.ID))

	body := gin.H{
		"name":     "Petr",
		"password": "123456",
		"role_id":  3,
	}

	req, _ := http.NewRequest("PUT", "/api/v1/employees/"+id, testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	employee = employessGetAll(t, engine, tokenOrg)[1]
	require.Equal(t, body["name"], employee.Name)
	require.Equal(t, body["role_id"], employee.RoleID)
}

func TestEmployeeDelete(t *testing.T) {
	engine, w := newController(t), httptest.NewRecorder()

	tokenOrg := orgGetToken(t, engine)
	tokenOwner := employeeGetOwnerToken(t, engine, tokenOrg)

	employees1 := employessGetAll(t, engine, tokenOrg) //автоматически созданный кассир
	employees1Len := len(employees1)

	id := strconv.Itoa(int(employees1[employees1Len-1].ID))

	req, _ := http.NewRequest("DELETE", "/api/v1/employees/"+id, nil)
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	employees2 := employessGetAll(t, engine, tokenOrg)
	employees2Len := len(employees2)

	require.NotEqual(t, employees1Len, employees2Len)
}
