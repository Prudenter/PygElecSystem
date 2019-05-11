package controllers

import "github.com/astaxie/beego"

/* 建立用户控制器类:UserController */

type UserController struct {
	beego.Controller
}

/* 定义函数负责用户注册页面展示 */
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}
