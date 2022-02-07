package root

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/iivkis/pos-ninja-backend/internal/myservice"
)

var baseURL string

func init() {
	baseURL = "http://localhost:8080/api/v1"
}

func url(location string) string {
	return baseURL + location
}

func marshal(body interface{}) *bytes.Reader {
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(b)
}

func doRequest(method string, location string, body interface{}) *http.Response {
	client := http.Client{}

	req, err := http.NewRequest(method, url(location), marshal(body))
	if err != nil {
		panic(err)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	return res
}

func unmarshal(res *http.Response) (data myservice.Response) {
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, &data); err != nil {
		panic(err)
	}
	return
}

func check(t *testing.T, data myservice.Response) {
	if !data.Status {
		t.Fatal(data)
	}
}

/* TESTS DATA*/

var newOrg = myservice.SignUpOrgInput{
	Name:     "OOO 'Ninja-Pro'",
	Email:    "test-ninj-alpha6@mail.ru",
	Password: "123456",
}

var orgToken string

/* TESTS */

func TestSignUpOrg(t *testing.T) {
	t.Run("main registation", func(t *testing.T) {
		req := doRequest("POST", "/auth/signUp.Org", newOrg)
		res := unmarshal(req)
		check(t, res)
	})
}

func TestSignInOrg(t *testing.T) {
	t.Run("main signIn", func(t *testing.T) {
		body := myservice.SignInOrgInput{
			Email:    newOrg.Email,
			Password: newOrg.Password,
		}
		req := doRequest("POST", "/auth/signIn.Org", body)
		res := unmarshal(req)
		check(t, res)
	})
}
