package app

//POS-Ninja Backend
//Created by Ivan Razmolodin (vk.com/ivan.razmolodin)

import (
	"log"

	controllerV1 "github.com/iivkis/pos.7-era.backend/internal/api/http/controllers/v1"
	"github.com/iivkis/pos.7-era.backend/internal/api/http/swagdocs"
	"github.com/iivkis/pos.7-era.backend/internal/components"
	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/internal/server"
)

func Launch() {
	log.Println("| SERVER LAUNCHING... |")

	//все компоненты необходимые для работы сервера
	components := components.New()

	// вешаем документацию
	swagdocs.Setup(components.Engine)

	// вешаем контроллеры
	{
		controllerV1.Setup(components)
	}

	// создаем сервер и запускаем
	{
		serv, servErr := server.NewServer(components.Engine)

		serv.Run("", *config.Flags.Port)
		log.Println("| SERVER UP |")

		panic(<-servErr)
	}
}
