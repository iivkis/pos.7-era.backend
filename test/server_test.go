package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

const baseURI = "http://localhost:80/api/v1/"

type BaseAPIResponse struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
}

var (
	tokens struct {
		Org  string
		Empl string
	}

	sessionID   uint
	orderInfoID uint

	testcfg struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Pin      string `json:"pin"`
	}
)

func init() {
	//load test config
	{
		f, err := os.OpenFile("./config.json", os.O_RDONLY, os.ModePerm)
		checkErr(err)

		b, err := io.ReadAll(f)
		checkErr(err)

		err = json.Unmarshal(b, &testcfg)
		checkErr(err)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func checkStatus(bar *BaseAPIResponse) {
	if !bar.Status {
		fmt.Println(bar.Data)
		panic(bar.Data)
	}
}

func marshal(data map[string]interface{}) io.Reader {
	b, err := json.Marshal(data)
	checkErr(err)
	return bytes.NewReader(b)
}

func readAll(resp *http.Response) []byte {
	b, err := io.ReadAll(resp.Body)
	checkErr(err)
	resp.Body.Close()

	return b
}

func unmarshal(res *http.Response) *BaseAPIResponse {
	var data BaseAPIResponse
	if err := json.Unmarshal(readAll(res), &data); err != nil {
		panic(err)
	}
	return &data
}

//tests
func TestSignIn(t *testing.T) {
	t.Run("org auth", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+"auth/signIn.Org", marshal(map[string]interface{}{
			"email":    testcfg.Email,
			"password": testcfg.Password,
		}))
		checkErr(err)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)

		data := decode.Data.(map[string]interface{})
		{
			tokens.Org = data["token"].(string)
		}
	})

	t.Run("employee auth", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+"auth/signIn.Employee", marshal(map[string]interface{}{
			"id":       1,
			"password": testcfg.Pin,
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Org)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)

		data := decode.Data.(map[string]interface{})
		{
			tokens.Empl = data["token"].(string)
			fmt.Println(tokens.Empl)
		}
	})
}

func TestSessionOpen(t *testing.T) {
	fmt.Println("Session open testing...")

	var idx uint
	t.Run("product create", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+"sessions", marshal(map[string]interface{}{
			"action": "open",
			"cash":   1000,
			"date":   123456789,
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)

		data := decode.Data.(map[string]interface{})
		{
			idx = uint(data["id"].(float64))
			sessionID = idx
		}
	})

	fmt.Println(idx)
	fmt.Println("")
}

func TestProduct(t *testing.T) {
	fmt.Println("Product testing...")

	var idx uint
	t.Run("product create", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+"products", marshal(map[string]interface{}{
			"amount":           1,
			"barcode":          3428043244,
			"category_id":      3,
			"name":             "string",
			"photo_id":         "string",
			"price":            10,
			"product_name_kkt": "string",
			"seller_percent":   10,
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)

		data := decode.Data.(map[string]interface{})
		{
			idx = uint(data["id"].(float64))
		}
	})

	t.Run("product get", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseURI+"products", nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	t.Run("product put", func(t *testing.T) {
		req, err := http.NewRequest("PUT", baseURI+fmt.Sprintf("%s/%d", "products", idx), marshal(map[string]interface{}{
			"amount":           2,
			"barcode":          2,
			"category_id":      1,
			"name":             "string1",
			"photo_id":         "string1",
			"price":            11,
			"product_name_kkt": "string1",
			"seller_percent":   11,
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	t.Run("product delete", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", baseURI+fmt.Sprintf("%s/%d", "products", idx), nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	fmt.Println(idx)
	fmt.Println("")
}

func TestCategories(t *testing.T) {
	fmt.Println("Categories testing...")

	var idx uint
	t.Run("category create", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+"categories", marshal(map[string]interface{}{
			"name": "string",
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)

		data := decode.Data.(map[string]interface{})
		{
			idx = uint(data["id"].(float64))
		}
	})

	t.Run("category get", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseURI+"categories", nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	t.Run("category put", func(t *testing.T) {
		req, err := http.NewRequest("PUT", baseURI+fmt.Sprintf("%s/%d", "categories", idx), marshal(map[string]interface{}{
			"name": "string1",
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	t.Run("category delete", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", baseURI+fmt.Sprintf("%s/%d", "categories", idx), nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	fmt.Println(idx)
	fmt.Println("")
}

func TestIngredients(t *testing.T) {
	fmt.Println("Ingredients testing...")

	var idx uint
	t.Run("ingredient create", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+"ingredients", marshal(map[string]interface{}{
			"count":          10,
			"measure_unit":   2,
			"name":           "string",
			"purchase_price": 10,
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)

		data := decode.Data.(map[string]interface{})
		{
			idx = uint(data["id"].(float64))
		}
	})

	t.Run("ingredient get", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseURI+"ingredients", nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	t.Run("ingredient put", func(t *testing.T) {
		req, err := http.NewRequest("PUT", baseURI+fmt.Sprintf("%s/%d", "ingredients", idx), marshal(map[string]interface{}{
			"count":          5,
			"measure_unit":   1,
			"name":           "string1",
			"purchase_price": 11,
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	t.Run("ingredient delete", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", baseURI+fmt.Sprintf("%s/%d", "ingredients", idx), nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	fmt.Println(idx)
	fmt.Println("")
}

func TestOrderInfo(t *testing.T) {
	fmt.Println("OrderInfo testing...")

	var idx uint
	t.Run("orderInfo create", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+"orderInfo", marshal(map[string]interface{}{
			"date":          123456,
			"employee_name": "string",
			"pay_type":      1,
			"session_id":    sessionID,
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)

		data := decode.Data.(map[string]interface{})
		{
			idx = uint(data["id"].(float64))
			orderInfoID = idx
		}
	})

	t.Run("orderInfo get", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseURI+"orderInfo", nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	t.Run("orderInfo delete", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", baseURI+fmt.Sprintf("%s/%d", "orderInfo", idx), nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	t.Run("orderInfo recovery", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+fmt.Sprintf("%s/%d", "orderInfo", idx), nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	fmt.Println(idx)
	fmt.Println("")
}

func TestOrderList(t *testing.T) {
	fmt.Println("OrderList testing...")

	var idx uint
	t.Run("orderList create", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+"orderList", marshal(map[string]interface{}{
			"count":         2,
			"order_info_id": orderInfoID,
			"product_id":    2,
			"product_name":  "string1",
			"product_price": 100,
			"session_id":    sessionID,
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)

		data := decode.Data.(map[string]interface{})
		{
			idx = uint(data["id"].(float64))
			orderInfoID = idx
		}
	})

	t.Run("orderList get", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseURI+"orderList", nil)
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	fmt.Println(idx)
	fmt.Println("")
}

func TestSessionClose(t *testing.T) {
	fmt.Println("Session close testing...")

	t.Run("session close", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURI+"sessions", marshal(map[string]interface{}{
			"action":      "close",
			"cash":        1000,
			"cash_earned": 1000,
			"bank_earned": 1000,
			"date":        123456789,
		}))
		checkErr(err)

		req.Header.Set("Authorization", tokens.Empl)

		res, err := http.DefaultClient.Do(req)
		checkErr(err)

		decode := unmarshal(res)
		checkStatus(decode)
	})

	fmt.Println(sessionID)
	fmt.Println("")
}
