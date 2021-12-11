package main

import (
	"github.com/iivkis/pos-ninja-backend/internal/app"
	_ "github.com/iivkis/pos-ninja-backend/internal/config"
)

//@BasePath /api/v1
//@Title POS-Ninja Backend API
//@Contact.Email razmolodinivan@mail.ru
//@Version 0.1-alpha

func main() {
	app.Launch()
}
