package routers

import (
	"PygElecSystem/PygElecSystem/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	//定义路由过滤器
	beego.InsertFilter("/user/*", beego.BeforeExec, filterFunc)
	beego.Router("/", &controllers.MainController{})
	//用户注册
	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister;post:HandleRegister")
	//处理发送短信业务，页面ajax发送的post请求
	beego.Router("/sendMsg", &controllers.UserController{}, "post:HandleSendMsg")
	//邮箱激活
	beego.Router("/register-email", &controllers.UserController{}, "get:ShowRegisterEmail;post:HandleRegisterEmail")
	//激活用户
	beego.Router("/active", &controllers.UserController{}, "get:HandleActive")
	//用户登录
	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	//首页面
	beego.Router("/index", &controllers.GoodsController{}, "get:ShowIndex")
	//退出登录
	beego.Router("/user/logout", &controllers.UserController{}, "get:Logout")
	//用户中心个人信息页面展示
	beego.Router("/user/userCenterInfo", &controllers.UserController{}, "get:ShowUserCenterInfo")
	//用户中心收货地址页面
	beego.Router("/user/site", &controllers.UserController{}, "get:ShowSite;post:HandleSite")
	//生鲜首页
	beego.Router("/index_sx", &controllers.GoodsController{}, "get:ShowIndex_sx")
	//商品详情
	beego.Router("/goodsDetail", &controllers.GoodsController{}, "get:ShowGoodsDetail")
	//展示同一类型所有商品
	beego.Router("/goodsType", &controllers.GoodsController{}, "get:ShowTypeList")
	//商品搜索
	beego.Router("/search", &controllers.GoodsController{}, "post:HandleSearch")
	//添加商品到购物车
	beego.Router("/addCart", &controllers.CartController{}, "post:HandleAddCart")
	//购物车展示
	beego.Router("/user/showCart", &controllers.CartController{}, "get:ShowCart")
	//修改购物车数量,包括:+,-,直接在输入框输入
	beego.Router("/changeCartCount", &controllers.CartController{}, "post:HandleChangeCartCount")
	//删除购物车商品
	beego.Router("/deleteCart", &controllers.CartController{}, "get:HandleDeleteCart")
	//添加商品到订单
	beego.Router("/user/addOrder", &controllers.OrderController{}, "post:ShowOrder")
	//提交订单
	beego.Router("/commitOrder", &controllers.OrderController{}, "post:HandleCommitOrder")
	//用户中心全部订单页面
	beego.Router("/user/userOrder", &controllers.UserController{}, "get:ShowUserOrder")
}

/* 定义过滤函数 */
func filterFunc(ctx *context.Context) {
	//过滤校验
	userName := ctx.Input.Session("userName")
	if userName == nil {
		ctx.Redirect(302, "/login")
		return
	}
}
