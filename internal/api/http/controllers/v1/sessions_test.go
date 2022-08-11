package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func sessionsOpen(t *testing.T, engine *gin.Engine, token string) (data sessionsActionResponse) {
	w := httptest.NewRecorder()

	body := gin.H{
		"action": "open",
		"date":   time.Now().UnixMilli(),
		"cash":   float64(testutil.RandomInt(1, 10000)),
	}

	req, _ := http.NewRequest("POST", basepath+"/sessions", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	require.NotEmpty(t, data.ID)
	require.NotEmpty(t, data.EmployeeID)

	return
}

func sessionsClose(t *testing.T, engine *gin.Engine, token string) (data sessionsActionResponse) {
	w := httptest.NewRecorder()

	body := gin.H{
		"action": "close",
		"date":   time.Now().UnixMilli(),
		"cash":   float64(testutil.RandomInt(1, 10000)),
	}

	req, _ := http.NewRequest("POST", basepath+"/sessions", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	require.NotEqual(t, 0, data.EmployeeID)

	return
}

func sessionsGetAll(t *testing.T, engine *gin.Engine, token string) (data sessionsGetAllResponse) {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", basepath+"/sessions", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)
	return
}

func TestSessionOpen(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	sessionsOpen(t, engine, tokenOwner)
}

func TestSessionClose(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	sessionID := sessionsOpen(t, engine, tokenOwner).ID

	var n = 10
	for i := 0; i < n; i++ {
		orderListCreate(t, engine, tokenOwner, sessionID)
	}

	sessionsClose(t, engine, tokenOwner)

	calc := orderListCalculation(t, engine, tokenOwner)
	sess := sessionsGetAll(t, engine, tokenOwner)

	require.Equal(t, calc.Total, sess[0].BankEarned+sess[0].CashEarned)
}

func TestSessionsGetAll(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	sessionID := sessionsOpen(t, engine, tokenOwner).ID

	var n = 5
	for i := 0; i < n; i++ {
		orderListCreate(t, engine, tokenOwner, sessionID)
	}

	sessionsClose(t, engine, tokenOwner)

	session := sessionsGetAll(t, engine, tokenOwner)[0]
	require.NotEmpty(t, session.ID)
	require.NotEmpty(t, session.CashEarned)
	require.NotEmpty(t, session.BankEarned)
	require.NotEmpty(t, session.CashClose)
	require.NotEmpty(t, session.CashOpen)
	require.NotEmpty(t, session.DateOpen)
	require.NotEmpty(t, session.DateClose)
}
