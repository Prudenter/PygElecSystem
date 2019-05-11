package routers

import (
	"PygElecSystem/PygElecSystem/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	//展示用户注册页面
	beego.Router("/register",&controllers.UserController{},"get:ShowRegister")

}
