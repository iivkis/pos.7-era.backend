package testutil

import (
	"bytes"
	"encoding/json"
	"io"
)

func Marshal(m any) io.Reader {
	d, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(d)
}

func Unmarshal(d *bytes.Buffer, m any) {
	b, err := io.ReadAll(d)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, m); err != nil {
		panic(err)
	}
}
