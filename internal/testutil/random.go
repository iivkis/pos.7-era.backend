package testutil

import (
	"math/rand"
	"time"
)

const alphabet = "qwertyuiopasdfghjklzxcvbnm"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString(n int) string {
	s := make([]byte, n)
	for i := range s {
		s[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(s)
}

// число пренадлежит интервалу [min; max]
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}
