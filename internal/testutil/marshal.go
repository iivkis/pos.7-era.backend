package testutil

import (
	"bytes"
	"encoding/json"
	"io"
)

func Marshal(m interface{}) io.Reader {
	d, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(d)
}
