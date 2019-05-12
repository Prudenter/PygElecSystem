package controllers

import "github.com/astaxie/beego"

type GoodsController struct {
	beego.Controller
}

/* 定义函数,负责商品首页展示 */
func (this *GoodsController) ShowIndex() {
	//获取session,用于页面显示
	userName := this.GetSession("userName")
	if userName != nil{
		this.Data["userName"] = userName.(string)
	}else {
		this.Data["userName"] = ""
	}
	this.TplName = "index.html"
}
