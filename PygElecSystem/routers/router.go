package routers

import (
	"PygElecSystem/PygElecSystem/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	//用户注册
	beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
	//处理发送短信业务，页面ajax发送的post请求
	beego.Router("/sendMsg",&controllers.UserController{},"post:HandleSendMsg")
	//邮箱激活
	beego.Router("/register-email",&controllers.UserController{},"get:ShowRegisterEmail;post:HandleRegisterEmail")
	//激活用户
	beego.Router("/active",&controllers.UserController{},"get:HandleActive")
	//用户登录
	beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")
	//首页面
	beego.Router("/index",&controllers.UserController{},"get:ShowIndex")
}
