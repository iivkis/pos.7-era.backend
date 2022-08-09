package controller

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/components"
	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/stretchr/testify/require"
)

const basepath = "/api/v1"

func newController(t *testing.T) *gin.Engine {
	config.Load("./../../../../../")
	c := Setup(components.New())
	return c.Engine
}

func TestAddController(t *testing.T) {
	engine := newController(t)
	require.NotNil(t, engine)
}
