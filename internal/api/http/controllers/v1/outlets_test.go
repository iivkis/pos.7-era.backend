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

func outletsGetAll(t *testing.T, engine *gin.Engine, token string) (data outletsGetAllResponse) {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", basepath+"/outlets", nil)
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	return
}

func outletsCreate(t *testing.T, engine *gin.Engine, token string) (data DefaultOutputModel) {
	w := httptest.NewRecorder()

	body := gin.H{
		"name": testutil.RandomString(50),
	}

	req, _ := http.NewRequest("POST", basepath+"/outlets", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, token)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var response Response
	testutil.Unmarshal(w.Body, &response)
	mapstructure.Decode(response.Data, &data)

	require.NotEmpty(t, data.ID)
	return data
}

func TestOutletsCreate(t *testing.T) {
	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	wg, n := new(sync.WaitGroup), 5

	wg.Add(n)
	defer wg.Wait()

	for i := 0; i < n; i++ {
		go func() {
			outletsCreate(t, engine, tokenOwner)
			wg.Done()
		}()
	}
}

func TestOutletsGetAll(t *testing.T) {
	engine := newController(t)
	tokenOrg := orgGetToken(t, engine)
	outletsGetAll(t, engine, tokenOrg)
}

func TestOutletsUpdate(t *testing.T) {
	engine, w := newController(t), httptest.NewRecorder()

	tokenOrg := orgGetToken(t, engine)
	tokenOwner := employeeGetOwnerToken(t, engine, tokenOrg)

	newOutlet := outletsCreate(t, engine, tokenOwner)
	outletID := strconv.Itoa(int(newOutlet.ID))

	body := gin.H{
		"name": testutil.RandomString(50),
	}

	req, _ := http.NewRequest("PUT", basepath+"/outlets/"+outletID, testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	outlets := outletsGetAll(t, engine, tokenOrg)
	outlet := outlets[len(outlets)-1] //берем последнюю добавленную
	require.Equal(t, outlet.Name, body["name"])
}

func TestOutletsDelete(t *testing.T) {
	engine, w := newController(t), httptest.NewRecorder()

	tokenOrg := orgGetToken(t, engine)
	tokenOwner := employeeGetOwnerToken(t, engine, tokenOrg)

	newOutlet := outletsCreate(t, engine, tokenOwner)
	outletID := strconv.Itoa(int(newOutlet.ID))

	outlets := outletsGetAll(t, engine, tokenOrg)
	require.Equal(t, 2, len(outlets))

	req, _ := http.NewRequest("DELETE", basepath+"/outlets/"+outletID, nil)
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	outlets = outletsGetAll(t, engine, tokenOrg)
	require.Equal(t, 1, len(outlets))
}
