package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestOrdersCreate(t *testing.T) {
	w := httptest.NewRecorder()

	engine := newController(t)
	tokenOwner := employeeGetOwnerToken(t, engine, orgGetToken(t, engine))

	productsCreate(t, engine, tokenOwner)
	product := productsGetAll(t, engine, tokenOwner)[0]

	fmt.Print("1")

	body := gin.H{
		"info": gin.H{
			"session_id":    sessionsOpen(t, engine, tokenOwner).ID,
			"employee_name": testutil.RandomString(10),
			"pay_type":      1,
			"discount":      10,
			"date":          time.Now().UnixMilli(),
		},

		"list": []gin.H{
			{
				"product_id":    product.ID,
				"product_name":  product.Name,
				"product_price": 100,
				"count":         3,
			},
			{
				"product_id":    product.ID,
				"product_name":  product.Name,
				"product_price": 200,
				"count":         3,
			},
		},
	}

	req, _ := http.NewRequest("POST", basepath+"/orders", testutil.Marshal(body))
	testutil.SetAuthorizationHeader(req, tokenOwner)

	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}
