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

func categoriesGetAll(t *testing.T, engine *gin.Engine, token string) categoriesGetAllResponse {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", basepath+"/categories", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var (
		response Response
		data     categoriesGetAllResponse
	)

	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	return data
}

func categoriesCreate(t *testing.T, engine *gin.Engine, token string) DefaultOutputModel {
	w := httptest.NewRecorder()

	body := gin.H{
		"name": testutil.RandomString(50),
	}

	req, _ := http.NewRequest("POST", basepath+"/categories", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var (
		response Response
		data     DefaultOutputModel
	)

	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	require.NotEmpty(t, data.ID)

	categories := categoriesGetAll(t, engine, token)
	require.NotEqual(t, len(categories), 0)

	return data
}

func TestCategoriesCreate(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	wg, n := new(sync.WaitGroup), 5

	wg.Add(n)
	defer wg.Wait()

	for i := 0; i < n; i++ {
		go func() {
			categoriesCreate(t, engine, tokenOwner)
			wg.Done()
		}()
	}
}

func TestCategoriesGetAll(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))
	categoriesGetAll(t, engine, tokenOwner)
}

func TestCategoriesUpdate(t *testing.T) {
	engine, w := newController(t), httptest.NewRecorder()

	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	newCategory := categoriesCreate(t, engine, tokenOwner)
	categoryID := strconv.Itoa(int(newCategory.ID))

	body := gin.H{
		"name": testutil.RandomString(10),
	}

	req, _ := http.NewRequest("PUT", basepath+"/categories/"+categoryID, testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	category := categoriesGetAll(t, engine, tokenOwner)[0]
	require.Equal(t, category.Name, body["name"])
}

func TestCategoriesDelete(t *testing.T) {
	engine, w := newController(t), httptest.NewRecorder()

	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	newCategory := categoriesCreate(t, engine, tokenOwner)
	categoryID := strconv.Itoa(int(newCategory.ID))

	categories := categoriesGetAll(t, engine, tokenOwner)
	require.NotEqual(t, len(categories), 0)

	req, _ := http.NewRequest("DELETE", basepath+"/categories/"+categoryID, nil)
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	categories = categoriesGetAll(t, engine, tokenOwner)
	require.Equal(t, len(categories), 0)
}
