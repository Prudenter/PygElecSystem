package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
)

/* 定义购物车控制器 */
type CartController struct {
	beego.Controller
}

/* 定义函数,负责添加商品到购物车业务处理 */
func (this *CartController) HandleAddCart() {
	//获取数据
	goodsId,err1 := this.GetInt("goodsId")
	num,err2 := this.GetInt("num")
	//返回ajax数据步骤
	//定义一个map容器
	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller,resp)
	if err1!=nil || err2 !=nil{
		resp["errno"] = 1
		resp["errmsg"] = "输入数据不完整!"
		return
	}
	//校验登录状态,获取当前登录用户名
	name := this.GetSession("userName")
	if name ==nil {
		resp["errno"] = 2
		resp["errmsg"] = "当前用户未登录,不能添加到购物车!"
		return
	}
	//处理数据
	//把购物车信息存储到redis的hash中
	conn,err := redis.Dial("tcp","127.0.0.1:6379")
	if err != nil {
		resp["errno"] = 3
		resp["errmsg"] = "服务器异常!"
		return
	}
	defer conn.Close()

	//先获取redis数据库中是否已添加了此件商品
	oldNum,err := redis.Int(conn.Do("hget","cart_"+name.(string),goodsId))
	//将购物车信息添加到redis中
	_,err = conn.Do("hset","cart_"+name.(string),goodsId,num+oldNum)

	if err != nil {
		resp["errno"] = 4
		resp["errmsg"] = "添加商品到服务器失败!"
		return
	}
	//返回数据
	resp["errno"] = 5
	resp["errmsg"] = "OK!"
}

/* 定义函数,负责购物车页面的展示 */
func (this *CartController) ShowCart() {
	//获取数据

	//校验数据

	//处理数据

	//返回数据
	this.TplName = "cart.html"
}