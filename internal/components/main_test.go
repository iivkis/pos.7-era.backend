package components

import (
	"testing"

	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	config.Load("./../../")

	components := New()

	require.NotEmpty(t, components.Engine)
	require.NotEmpty(t, components.Postman)
	require.NotEmpty(t, components.Repo)
	require.NotEmpty(t, components.S3cloud)
	require.NotEmpty(t, components.Strcode)
	require.NotEmpty(t, components.TokenMaker)
}
