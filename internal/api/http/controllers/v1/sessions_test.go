package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/stretchr/testify/require"
)

func sessionsOpen(t *testing.T, engine *gin.Engine, token string) sessionsOpenOrCloseResponse {
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

}

func TestSessionOpen(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	sessionsOpen(t, engine, tokenOwner)
}
