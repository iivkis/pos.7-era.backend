package app

import (
	"fmt"
	"time"

	"github.com/iivkis/pos-ninja-backend/internal/config"
	"github.com/iivkis/pos-ninja-backend/internal/handler"
	"github.com/iivkis/pos-ninja-backend/internal/myservice"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"github.com/iivkis/pos-ninja-backend/internal/server"
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"github.com/iivkis/pos-ninja-backend/pkg/mailagent"
	"github.com/iivkis/strcode"
)

func Launch() {
	welcomInfo()
	fmt.Println("Server launching... \\-0.0-/")

	//pkg
	_authjwt := authjwt.NewAuthJWT([]byte(config.Env.Secret))

	_strcode, err := strcode.NewStrcode(config.Env.Secret, ":", time.Hour*24)
	if err != nil {
		panic(err)
	}

	_mailagent := mailagent.NewMailAgent(config.Env.EmailLogin, config.Env.EmailPwd)
	_mailagent.LoadTemplatesFromDir(config.File.EmailTmplDir)
	if err := _mailagent.SendTemplate(config.Env.EmailForNotify, "service_run.html", mailagent.Value{
		"name":     config.Env.ServerName,
		"protocol": config.Env.Protocol,
		"host":     config.Env.Host,
		"port":     config.Env.Port,
	}); err != nil {
		panic(err)
	}

	//internal
	_repo := repository.NewRepository(_authjwt)
	_service := myservice.NewMyService(_repo, _strcode, _mailagent)
	_handler := handler.NewHttpHandler(_service)
	_server := server.NewServer(_handler)

	//run server
	var done = make(chan byte)
	go func() {
		if err := _server.Listen(config.Env.Host, config.Env.Port); err != nil {
			panic(err)
		}
	}()

	fmt.Print("Server launched =D\n\n")
	<-done
}

func welcomInfo() {
	fmt.Println("------------------------------------------------------")
	fmt.Println("[APP NAME] POS-Ninja-Backend (version: 0.1-alpha)")
	fmt.Println("[CREATED BY] Ivan Razmolodin (vk.com/ivan.razmolodin)")
	fmt.Println("------------------------------------------------------")
}
