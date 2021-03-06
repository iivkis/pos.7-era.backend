package app

//POS-Ninja Backend
//Created by Ivan Razmolodin (vk.com/ivan.razmolodin)

import (
	"fmt"
	"time"

	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/internal/handler"
	"github.com/iivkis/pos.7-era.backend/internal/myservice"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/selectelS3Cloud"
	"github.com/iivkis/pos.7-era.backend/internal/server"
	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
	"github.com/iivkis/pos.7-era.backend/pkg/mailagent"
	"github.com/iivkis/strcode"
)

func Launch() {
	fmt.Println("Server launching... \\-0.0-/")

	//pkg
	_authjwt := authjwt.NewAuthJWT([]byte(config.Env.Secret))

	_strcode, err := strcode.NewStrcode(config.Env.Secret, ":", time.Hour*24)
	if err != nil {
		panic(err)
	}

	_mailagent := mailagent.NewMailAgent(config.Env.EmailLogin, config.Env.EmailPwd)

	//internal
	_s3cloud := selectelS3Cloud.NewSelectelS3Cloud(config.Env.SelectelS3AccessKey, config.Env.SelectelS3SecretKey, "https://cb027f6f-0eed-40c8-8f6a-7fbc35d7224b.selcdn.net")
	_repo := repository.NewRepository(_authjwt)
	_service := myservice.NewMyService(_repo, _strcode, _mailagent, _authjwt, _s3cloud)
	_handler := handler.NewHttpHandler(_service)
	_server := server.NewServer(_handler)

	//run server
	var done = make(chan string)
	go func() {
		if err := _server.Listen(*config.Flags.Port); err != nil {
			done <- err.Error()
		}
	}()

	fmt.Print("Server launched =D\n\n")
	panic(<-done)
}
