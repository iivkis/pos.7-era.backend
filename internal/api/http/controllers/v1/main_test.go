package controller

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/s3cloud"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"github.com/iivkis/pos.7-era.backend/pkg/mailagent"
	"github.com/iivkis/strcode"
	"github.com/stretchr/testify/require"
)

const basepath = "/api/v1"

func newController(t *testing.T) *gin.Engine {
	config.Load("./../../../../../")

	s3cloud := s3cloud.NewSelectelS3Cloud(config.Env.SelectelS3AccessKey, config.Env.SelectelS3SecretKey, "https://cb027f6f-0eed-40c8-8f6a-7fbc35d7224b.selcdn.net")
	postman := mailagent.NewMailAgent(config.Env.EmailLogin, config.Env.EmailPassword)
	strcode, _ := strcode.NewStrcode(config.Env.TokenSecretKey, ":", time.Second*90)
	tokenMaker := authjwt.NewAuthJWT([]byte(config.Env.TokenSecretKey))
	repo := repository.NewRepository(tokenMaker)

	c := AddController(gin.Default(), repo, strcode, postman, tokenMaker, s3cloud)
	return c.Engine
}

func TestAddController(t *testing.T) {
	engine := newController(t)
	require.NotNil(t, engine)
}
