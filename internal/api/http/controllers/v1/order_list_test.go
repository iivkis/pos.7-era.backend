package controller

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func orderListGetAll(t *testing.T, engine *gin.Engine, token string) (data orderListGetAllResponse) {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", basepath+"/orderList", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	return
}

func orderListCreate(t *testing.T, engine *gin.Engine, token string, sessionID uint) (data DefaultOutputModel) {
	w := httptest.NewRecorder()

	productID := productsCreate(t, engine, token).ID
	orderInfoID := orderInfoCreate(t, engine, token, sessionID).ID

	body := gin.H{
		"product_id":    productID,
		"order_info_id": orderInfoID,
		"session_id":    sessionID,

		"count":         testutil.RandomInt(1, 100),
		"product_name":  testutil.RandomString(10),
		"product_price": rand.Float64() * 100,
	}

	req, _ := http.NewRequest("POST", basepath+"/orderList", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)
	log.Println(response)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotEqual(t, data.ID, 0)

	orderList := orderListGetAll(t, engine, token)[0]
	require.Equal(t, orderList.ProductID, body["product_id"])
	require.Equal(t, orderList.OrderInfoID, body["order_info_id"])
	require.Equal(t, orderList.SessionID, body["session_id"])
	require.Equal(t, orderList.Count, body["count"])
	require.Equal(t, orderList.ProductPrice, body["product_price"])
	require.Equal(t, orderList.ProductName, body["product_name"])

	return
}

func TestOrderListGetAll(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	orderListGetAll(t, engine, tokenOwner)
}

func TestOrderListCreate(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	sessionID := sessionsOpen(t, engine, tokenOwner).ID
	orderListCreate(t, engine, tokenOwner, sessionID)
}
