package testutil

import "net/http"

func SetAuthorizationHeader(req *http.Request, token string) {
	req.Header.Set("Authorization", token)
}
