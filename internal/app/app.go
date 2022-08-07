package app

//POS-Ninja Backend
//Created by Ivan Razmolodin (vk.com/ivan.razmolodin)

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	controllerV1 "github.com/iivkis/pos.7-era.backend/internal/api/http/controllers/v1"
	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/s3cloud"
	"github.com/iivkis/pos.7-era.backend/internal/server"
	"github.com/iivkis/pos.7-era.backend/internal/servutil"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"github.com/iivkis/pos.7-era.backend/pkg/mailagent"
	"github.com/iivkis/strcode"
)

func Launch() {
	log.Println("| SERVER LAUNCHING... |")

	ENGINE := gin.Default()

	tokenMaker := authjwt.NewAuthJWT([]byte(config.Env.TokenSecretKey))

	strcode, err := strcode.NewStrcode(config.Env.TokenSecretKey, ":", time.Hour*24)
	servutil.PanicIfErr(err)

	postman := mailagent.NewMailAgent(config.Env.EmailLogin, config.Env.EmailPassword)

	s3cloud := s3cloud.NewSelectelS3Cloud(config.Env.SelectelS3AccessKey, config.Env.SelectelS3SecretKey, "https://cb027f6f-0eed-40c8-8f6a-7fbc35d7224b.selcdn.net")
	repo := repository.NewRepository(tokenMaker)

	//setup routs
	{
		controllerV1.AddController(ENGINE, repo, strcode, postman, tokenMaker, s3cloud)
	}

	var done = make(chan string)
	go func() {
		serv := server.NewServer(ENGINE)

		if err := serv.Run("", *config.Flags.Port); err != nil {
			done <- err.Error()
		}
	}()

	log.Println("| SERVER UP |")
	panic(<-done)
}
