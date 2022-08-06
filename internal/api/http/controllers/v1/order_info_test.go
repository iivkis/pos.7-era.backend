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

func orderInfoGetAll(t *testing.T, engine *gin.Engine, token string) (data orderInfoGetAllResponse) {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", basepath+"/orderInfo", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	return
}

func orderInfoCreate(t *testing.T, engine *gin.Engine, token string) (data DefaultOutputModel) {
	w := httptest.NewRecorder()

	session := sessionsOpen(t, engine, token)

	log.Println(session)

	body := gin.H{
		"session_id":    session.ID,
		"pay_type":      testutil.RandomInt(0, 2),
		"employee_name": testutil.RandomString(20),
		"date":          time.Now().UnixMilli(),
	}

	req, _ := http.NewRequest("POST", basepath+"/orderInfo", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)
	require.NotEmpty(t, data.ID)
	return
}

func TestOrderInfoCreate(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	orderInfoCreate(t, engine, tokenOwner)
}

func TestOrderInfoGetAll(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	orderInfoGetAll(t, engine, tokenOwner)
}
