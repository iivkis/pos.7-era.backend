package app

import (
	"fmt"

	"github.com/iivkis/pos-ninja-backend/internal/handler"
	"github.com/iivkis/pos-ninja-backend/internal/myservice"
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"github.com/iivkis/pos-ninja-backend/internal/server"
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
)

func Launch() {
	welcomInfo()
	fmt.Println("Server launching... \\-0.0-/")

	//pkg
	_authjwt := authjwt.NewAuthJWT([]byte("key12345"))

	//internal
	_repo := repository.NewRepository(_authjwt)
	_service := myservice.NewMyService(_repo)
	_handler := handler.NewHttpHandler(_service)
	_server := server.NewServer(_handler)

	//run server
	var done = make(chan byte)
	go func() {
		if err := _server.Listen(); err != nil {
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
