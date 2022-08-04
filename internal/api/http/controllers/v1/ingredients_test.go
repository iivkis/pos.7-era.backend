package controller

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func ingredientsGetAll(t *testing.T, engine *gin.Engine, token string) (data ingredientGetAllResponse) {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", basepath+"/ingredients", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	return
}

func ingredientsCreate(t *testing.T, engine *gin.Engine, token string) (data DefaultOutputModel) {
	w := httptest.NewRecorder()

	body := gin.H{
		"name":           testutil.RandomString(10),
		"count":          rand.Float64() * 20,
		"purchase_price": rand.Float64() * 100,
		"measure_unit":   1 + rand.Intn(2),
	}

	req, _ := http.NewRequest("POST", basepath+"/ingredients", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	require.NotEqual(t, data.ID, 0)

	ingredient := ingredientsGetAll(t, engine, token)[0]
	require.Equal(t, ingredient.Name, body["name"])
	require.Equal(t, ingredient.Count, body["count"])
	require.Equal(t, ingredient.PurchasePrice, body["purchase_price"])
	require.Equal(t, ingredient.MeasureUnit, body["measure_unit"])

	return
}

func TestIngredientsGetAll(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	ingredientsGetAll(t, engine, tokenOwner)
}

func TestIngredientsCreate(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	ingredientsCreate(t, engine, tokenOwner)
}

func TestIngredientsUpdate(t *testing.T) {
	engine := newController(t)
	w := httptest.NewRecorder()

	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	newIngredient := ingredientsCreate(t, engine, tokenOwner)
	ingredientID := strconv.Itoa(int(newIngredient.ID))

	body := gin.H{
		"name":           testutil.RandomString(10),
		"count":          rand.Float64() * 20,
		"purchase_price": rand.Float64() * 100,
		"measure_unit":   1 + rand.Intn(2),
	}

	req, _ := http.NewRequest("PUT", basepath+"/ingredients/"+ingredientID, testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)

	ingredient := ingredientsGetAll(t, engine, tokenOwner)[0]
	require.Equal(t, ingredient.Name, body["name"])
	require.Equal(t, ingredient.Count, body["count"])
	require.Equal(t, ingredient.PurchasePrice, body["purchase_price"])
	require.Equal(t, ingredient.MeasureUnit, body["measure_unit"])
}

func TestIngredientsDelete(t *testing.T) {
	engine, w := newController(t), httptest.NewRecorder()

	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	newIngredient := ingredientsCreate(t, engine, tokenOwner)
	ingredientID := strconv.Itoa(int(newIngredient.ID))

	ingredients := ingredientsGetAll(t, engine, tokenOwner)
	require.NotEqual(t, len(ingredients), 0)

	req, _ := http.NewRequest("DELETE", basepath+"/ingredients/"+ingredientID, nil)
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	ingredients = ingredientsGetAll(t, engine, tokenOwner)
	require.Equal(t, len(ingredients), 0)
}
