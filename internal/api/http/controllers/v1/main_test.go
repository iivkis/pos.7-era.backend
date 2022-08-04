package controller

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"github.com/iivkis/pos.7-era.backend/pkg/mailagent"
	"github.com/iivkis/strcode"
	"github.com/stretchr/testify/require"
)

const basepath = "/api/v1"

func newController(t *testing.T) *gin.Engine {
	config.Load("./../../../../../")

	postman := mailagent.NewMailAgent(config.Env.EmailLogin, config.Env.EmailPassword)
	strcode, _ := strcode.NewStrcode(config.Env.TokenSecretKey, ":", time.Second*90)
	tokenMaker := authjwt.NewAuthJWT([]byte(config.Env.TokenSecretKey))
	repo := repository.NewRepository(tokenMaker)

	c := AddController(gin.Default(), repo, strcode, postman, tokenMaker)
	return c.Engine
}

func TestAddController(t *testing.T) {
	engine := newController(t)
	require.NotNil(t, engine)
}
