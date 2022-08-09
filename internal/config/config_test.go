package config

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		Load("./../../")
		onceFiles = sync.Once{}
	})

	t.Run("bad", func(t *testing.T) {
		defer func() {
			require.NotEmpty(t, recover())
		}()

		Load(".")
		onceFiles = sync.Once{}
	})
}
