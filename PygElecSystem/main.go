package main

import (
	_ "PygElecSystem/PygElecSystem/routers"
	"github.com/astaxie/beego"
	_ "PygElecSystem/PygElecSystem/models"
)

func main() {
	beego.Run()
}

