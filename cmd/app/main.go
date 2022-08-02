package main

import (
	"github.com/iivkis/pos.7-era.backend/internal/app"
	_ "github.com/iivkis/pos.7-era.backend/internal/config"
)

//@BasePath /api/v1
//@Title POS-Ninja Backend API
//@Contact.Email razmolodinivan@mail.ru
//@Version 0.2.0

func main() {
	app.Launch()
}
