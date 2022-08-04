package controller

import (
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

	req, _ := http.NewRequest("GET", basepath+"/ordersList", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	return
}

func orderListCreate(t *testing.T, engine *gin.Engine, token string) (data DefaultOutputModel) {
	w := httptest.NewRecorder()

	product := productsCreate(t, engine, token)

	body := gin.H{
		"product_id":    product.ID,
		"order_info_id": 0,
		"session_id":    0,

		"count":         rand.Float64() * 20,
		"product_name":  testutil.RandomString(10),
		"product_price": rand.Float64() * 100,
	}

	req, _ := http.NewRequest("POST", basepath+"/orderList", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

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
	orderListCreate(t, engine, tokenOwner)
}

// func TestIngredientsUpdate(t *testing.T) {
// 	engine := newController(t)
// 	w := httptest.NewRecorder()

// 	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

// 	newIngredient := ingredientsCreate(t, engine, tokenOwner)
// 	ingredientID := strconv.Itoa(int(newIngredient.ID))

// 	body := gin.H{
// 		"name":           testutil.RandomString(10),
// 		"count":          rand.Float64() * 20,
// 		"purchase_price": rand.Float64() * 100,
// 		"measure_unit":   1 + rand.Intn(2),
// 	}

// 	req, _ := http.NewRequest("PUT", basepath+"/ingredients/"+ingredientID, testutil.Marshal(body))
// 	testutil.SetAuthorizationHeader(req, tokenOwner)

// 	engine.ServeHTTP(w, req)
// 	require.Equal(t, http.StatusOK, w.Code)

// 	var response Response
// 	testutil.Unmarshal(w.Body, &response)

// 	ingredient := ingredientsGetAll(t, engine, tokenOwner)[0]
// 	require.Equal(t, ingredient.Name, body["name"])
// 	require.Equal(t, ingredient.Count, body["count"])
// 	require.Equal(t, ingredient.PurchasePrice, body["purchase_price"])
// 	require.Equal(t, ingredient.MeasureUnit, body["measure_unit"])
// }

// func TestIngredientsDelete(t *testing.T) {
// 	engine, w := newController(t), httptest.NewRecorder()

// 	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

// 	newIngredient := ingredientsCreate(t, engine, tokenOwner)
// 	ingredientID := strconv.Itoa(int(newIngredient.ID))

// 	ingredients := ingredientsGetAll(t, engine, tokenOwner)
// 	require.NotEqual(t, len(ingredients), 0)

// 	req, _ := http.NewRequest("DELETE", basepath+"/ingredients/"+ingredientID, nil)
// 	testutil.SetAuthorizationHeader(req, tokenOwner)

// 	engine.ServeHTTP(w, req)
// 	require.Equal(t, http.StatusOK, w.Code)

// 	ingredients = ingredientsGetAll(t, engine, tokenOwner)
// 	require.Equal(t, len(ingredients), 0)
// }
