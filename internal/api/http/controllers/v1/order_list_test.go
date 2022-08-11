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

	productsCreate(t, engine, token)
	product := productsGetAll(t, engine, token)[0]
	orderInfoID := orderInfoCreate(t, engine, token, sessionID).ID

	body := gin.H{
		"product_id":    product.ID,
		"order_info_id": orderInfoID,
		"session_id":    sessionID,

		"count":         testutil.RandomInt(1, 100),
		"product_name":  testutil.RandomString(10),
		"product_price": product.Price,
	}

	req, _ := http.NewRequest("POST", basepath+"/orderList", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotEqual(t, data.ID, 0)

	orderList := orderListGetAll(t, engine, token)
	orderListLast := orderList[len(orderList)-1]

	require.Equal(t, orderListLast.ProductID, body["product_id"])
	require.Equal(t, orderListLast.OrderInfoID, body["order_info_id"])
	require.Equal(t, orderListLast.SessionID, body["session_id"])
	require.Equal(t, orderListLast.Count, body["count"])
	require.Equal(t, orderListLast.ProductPrice, body["product_price"])
	require.Equal(t, orderListLast.ProductName, body["product_name"])

	return
}

func orderListCalculation(t *testing.T, engine *gin.Engine, token string) (data orderListCalcResponse) {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", basepath+"/orderList.Calc", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

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

func TestOrderListCalculation(t *testing.T) {
	engine := newController(t)

	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	sessionID := sessionsOpen(t, engine, tokenOwner).ID

	orderListCreate(t, engine, tokenOwner, sessionID)
	orderListCreate(t, engine, tokenOwner, sessionID)

	orderInfoDelete(t, engine, tokenOwner, orderInfoGetAll(t, engine, tokenOwner)[0].ID) // удаляем второй чек

	orderList := orderListGetAll(t, engine, tokenOwner)
	calc := orderListCalculation(t, engine, tokenOwner)

	require.Equal(t, orderList[1].ProductPrice*float64(orderList[1].Count), calc.Total)
}
