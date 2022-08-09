package components

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/s3cloud"
	"github.com/iivkis/pos.7-era.backend/internal/tokenmaker"
	"github.com/iivkis/pos.7-era.backend/pkg/postman"
	"github.com/iivkis/strcode"
)

type Components struct {
	Engine     *gin.Engine
	Repo       *repository.Repository
	Strcode    *strcode.Strcode
	Postman    *postman.Postman
	TokenMaker *tokenmaker.TokenMaker
	S3cloud    *s3cloud.SelectelS3Cloud
}

func New() Components {
	engine := gin.Default()

	// отправка email писем
	postman := postman.NewMailAgent(config.Env.EmailLogin, config.Env.EmailPassword)

	// шифрование строки
	strcode, err := strcode.NewStrcode(config.Env.TokenSecretKey, ":", time.Hour*24)
	if err != nil {
		panic(err)
	}

	// токены авторизации для API
	tokenMaker := tokenmaker.NewTokenMaker([]byte(config.Env.TokenSecretKey))

	//репозиторий и бд
	repo := repository.NewRepository(tokenMaker)

	// облачное хранилище
	s3cloud := s3cloud.NewSelectelS3Cloud(config.Env.SelectelS3AccessKey, config.Env.SelectelS3SecretKey, "https://cb027f6f-0eed-40c8-8f6a-7fbc35d7224b.selcdn.net")

	return Components{
		Engine:     engine,
		TokenMaker: tokenMaker,
		Repo:       repo,
		Strcode:    strcode,
		Postman:    postman,
		S3cloud:    s3cloud,
	}
}
