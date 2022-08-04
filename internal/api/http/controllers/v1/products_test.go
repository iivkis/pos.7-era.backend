package controller

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func productsGetAll(t *testing.T, engine *gin.Engine, token string) (data productGetAllResponse) {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", basepath+"/products", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response

	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	return
}

func productsCreate(t *testing.T, engine *gin.Engine, token string) (data DefaultOutputModel) {
	w := httptest.NewRecorder()

	newCategory := categoriesCreate(t, engine, token)
	categoryID := newCategory.ID

	body := gin.H{
		"category_id":      categoryID,
		"name":             testutil.RandomString(50),
		"product_name_kkt": testutil.RandomString(50),
		"barcode":          testutil.RandomInt(10000, 100000),
		"amount":           testutil.RandomInt(1, 10000),
		"price":            float64(testutil.RandomInt(1, 10000)),
		"seller_percent":   float64(testutil.RandomInt(1, 100)),
	}

	req, _ := http.NewRequest("POST", basepath+"/products", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	require.NotEmpty(t, data.ID)
	return data
}

func TestProductssCreate(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	wg, n := new(sync.WaitGroup), 5

	wg.Add(n)
	defer wg.Wait()

	for i := 0; i < n; i++ {
		go func() {
			productsCreate(t, engine, tokenOwner)
			wg.Done()
		}()
	}
}

func TestProductsGetAll(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	productsGetAll(t, engine, tokenOwner)
}

func TestProductsUpdate(t *testing.T) {
	engine, w := newController(t), httptest.NewRecorder()

	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	newProduct := productsCreate(t, engine, tokenOwner)
	productID := strconv.Itoa(int(newProduct.ID))

	newCategory := categoriesCreate(t, engine, tokenOwner)
	categoryID := newCategory.ID

	body := gin.H{
		"category_id":      categoryID,
		"name":             testutil.RandomString(50),
		"product_name_kkt": testutil.RandomString(50),
		"barcode":          testutil.RandomInt(10000, 100000),
		"amount":           testutil.RandomInt(1, 10000),
		"price":            float64(testutil.RandomInt(1, 10000)),
		"seller_percent":   float64(testutil.RandomInt(1, 100)),
	}

	req, _ := http.NewRequest("PUT", basepath+"/products/"+productID, testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	product := productsGetAll(t, engine, tokenOwner)[0]
	require.Equal(t, product.CategoryID, body["category_id"])
	require.Equal(t, product.Name, body["name"])
	require.Equal(t, product.ProductNameKKT, body["product_name_kkt"])
	require.Equal(t, product.Barcode, body["barcode"])
	require.Equal(t, product.Amount, body["amount"])
	require.Equal(t, product.Price, body["price"])
	require.Equal(t, product.SellerPercent, body["seller_percent"])
}

func TestProductsDelete(t *testing.T) {
	engine, w := newController(t), httptest.NewRecorder()

	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	newProduct := productsCreate(t, engine, tokenOwner)
	productID := strconv.Itoa(int(newProduct.ID))

	products := productsGetAll(t, engine, tokenOwner)
	require.NotEqual(t, len(products), 0)

	req, _ := http.NewRequest("DELETE", basepath+"/products/"+productID, nil)
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	products = productsGetAll(t, engine, tokenOwner)
	require.Equal(t, len(products), 0)
}
