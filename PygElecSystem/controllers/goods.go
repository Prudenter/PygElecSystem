package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"strings"
	"PygElecSystem/PygElecSystem/models"
)

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

/* 定义函数,获取当前登录用户 */
func GoodsGetUser(this *GoodsController) models.User {
	//根据session获取当前登录用户名
	userName := this.GetSession("userName")
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user,"Name")
	//手机号码加密
	str := user.Phone
	user.Phone = strings.Join([]string{str[0:3],"****",str[7:]},"");
	return user
}

/* 定义函数,查询当前用户默认地址*/
func GoodsGetUserAddr(this *GoodsController) models.Address {
	//查询数据库,显示默认地址
	o := orm.NewOrm()
	var address models.Address
	//获取当前用户的默认地址
	userName := this.GetSession("userName").(string)
	qs := o.QueryTable("Address")
	qs.RelatedSel("User").Filter("User__Name", userName).Filter("IsDefault", true).One(&address)
	//手机号码加密
	if address.Phone != "" {
		str := address.Phone
		address.Phone = strings.Join([]string{str[0:3], "****", str[7:]}, "")
	}
	return address
}

/* 定义函数,负责全部订单页面显示 */
func (this *GoodsController) ShowOrder() {
	//调用函数,获取当前登录用户
	user := GoodsGetUser(this)
	this.Data["user"] = user
	////调用函数,获取当前登录用户的默认地址
	this.Data["address"] = GoodsGetUserAddr(this)
	//实现视图布局,将模板与主要部分连接其起来
	this.Layout = "user_center_layout.html"
	this.Data["num"] = 2
	this.TplName = "user_center_order.html"
}