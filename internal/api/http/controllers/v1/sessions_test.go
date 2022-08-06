package controller

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func sessionsOpen(t *testing.T, engine *gin.Engine, token string) (data sessionsOpenOrCloseResponse) {
	w := httptest.NewRecorder()

	body := gin.H{
		"action":      "open",
		"date":        time.Now().UnixMilli(),
		"cash":        float64(testutil.RandomInt(1, 10000)),
		"cash_earned": float64(testutil.RandomInt(1, 10000)),
		"bank_earned": float64(testutil.RandomInt(1, 10000)),
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

	log.Println(response)

	return
}

func sessionsClose(t *testing.T, engine *gin.Engine, token string) (data sessionsOpenOrCloseResponse) {
	w := httptest.NewRecorder()

	body := gin.H{
		"action":      "close",
		"date":        time.Now().UnixMilli(),
		"cash":        float64(testutil.RandomInt(1, 10000)),
		"cash_earned": float64(testutil.RandomInt(1, 10000)),
		"bank_earned": float64(testutil.RandomInt(1, 10000)),
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
	sessionsOpen(t, engine, tokenOwner)
	sessionsClose(t, engine, tokenOwner)
}

func TestSessionsGetAll(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	sessionsOpen(t, engine, tokenOwner)
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
