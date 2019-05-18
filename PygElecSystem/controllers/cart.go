package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	"fmt"
	"github.com/astaxie/beego/orm"
	"PygElecSystem/PygElecSystem/models"
)

/* 定义购物车控制器 */
type CartController struct {
	beego.Controller
}

/* 定义函数,负责添加商品到购物车业务处理 */
func (this *CartController) HandleAddCart() {
	//获取数据
	goodsId, err1 := this.GetInt("goodsId")
	num, err2 := this.GetInt("num")
	//返回ajax数据步骤
	//定义一个map容器
	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller, resp)
	if err1 != nil || err2 != nil {
		resp["errno"] = 1
		resp["errmsg"] = "输入数据不完整!"
		return
	}
	//校验登录状态,获取当前登录用户名
	name := this.GetSession("userName")
	if name == nil {
		resp["errno"] = 2
		resp["errmsg"] = "当前用户未登录,不能添加到购物车!"
		return
	}
	//处理数据
	//把购物车信息存储到redis的hash中
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		resp["errno"] = 3
		resp["errmsg"] = "服务器异常!"
		return
	}
	defer conn.Close()

	//先获取redis数据库中是否已添加了此件商品
	oldNum, err := redis.Int(conn.Do("hget", "cart_"+name.(string), goodsId))
	//将购物车信息添加到redis中
	_, err = conn.Do("hset", "cart_"+name.(string), goodsId, num+oldNum)

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
	//获取session中的用户名,根据用户名在redis数据库中获取当前用户的购物车数据
	userName := this.GetSession("userName")

	//连接redis
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Print("redis连接错误!")
		this.Redirect("/index_sx", 302)
		return
	}
	defer conn.Close()
	//查询所有购物车数据,返回的是数字字符对应的ASCII码值,类型是interface{}
	resp, err := conn.Do("hgetall", "cart_"+userName.(string))
	if err != nil {
		fmt.Println("获取数据异常!")
		this.Redirect("/index_sx", 302)
		return
	}
	//将interface{}转换为int类型切片
	result, _ := redis.Ints(resp, err)

	//处理数据
	//定义大容器,存入购物车中所有的商品信息
	var goods []map[string]interface{}
	o := orm.NewOrm()
	//定义商品变量
	var goodsSKU models.GoodsSKU
	//定义商品小计,总价和总件数
	var littlePrice, totalPrice, totalCount int
	for i := 0; i < len(result); i += 2 {
		//定义行容器
		temp := make(map[string]interface{})
		//注意加入购物车时是以hash类型存入的,其中key = userNname,field= result[i] = goodsId,value = result[i+1]
		goodsSKU.Id = result[i]
		o.Read(&goodsSKU)
		//给行容器赋值
		temp["goodsSKU"] = goodsSKU
		temp["count"] = result[i+1]

		//计算小计,总价和总件数
		littlePrice = result[i+1] * goodsSKU.Price
		temp["littlePrice"] = littlePrice
		totalPrice += littlePrice
		totalCount ++

		//把行容器加入总容器中
		goods = append(goods, temp)
	}

	//返回数据
	this.Data["totalPrice"] = totalPrice
	this.Data["totalCount"] = totalCount
	this.Data["goods"] = goods
	this.TplName = "cart.html"
}

/* 定义函数,负责修改购物车数量的业务处理 */
func (this *CartController) HandleChangeCartCount() {
	//获取数据
	//获取商品id
	goodsId,err1 := this.GetInt("goodsId")
	//获取数量
	count,err2 := this.GetInt("count")

	//校验数据,返回json数据给ajax
	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller,resp)
	if err1 != nil || err2 != nil {
		fmt.Println(err1,err2)
		fmt.Println("获取数据不完整!")
		resp["errno"] = 1
		resp["errmsg"] = "获取数据不完整!"
		return
	}
	//获取session中的用户名
	userName := this.GetSession("userName")
	if userName==nil {
		fmt.Println("当前用户未登录!")
		resp["errno"] = 2
		resp["errmsg"] = "当前用户未登录!"
		return
	}

	//修改redis中当前用户当前商品的数量
	//连接redis数据库
	conn,err := redis.Dial("tcp","127.0.0.1:6379")
	if err != nil {
		fmt.Println("redis连接异常!")
		resp["errno"] = 3
		resp["errmsg"] = "网络连接异常!"
		return
	}
	//修改数据
	_,err = conn.Do("hset","cart_"+userName.(string),goodsId,count)
	if err!=nil {
		fmt.Println("redis写入异常!")
		resp["errno"] = 4
		resp["errmsg"] = "redis写入异常!"
		return
	}
	//返回数据
	resp["errno"] = 5
	resp["errmsg"] = "ok!"
}

/* 定义函数,负责删除购物车商品 */
func (this *CartController) HandleDeleteCart() {
	//获取数据
	id,err := this.GetInt("id")
	//校验数据
	if err != nil {
		fmt.Println("删除连接错误!")
		this.Redirect("/user/showCart",302)
		return
	}
	//校验是否是登录状态
	userName := this.GetSession("userName")
	if userName == nil {
		fmt.Println("当前用户不在登录状态!")
		this.Redirect("/user/showCart",302)
		return
	}

	//向redis中写入数据
	conn,err :=redis.Dial("tcp","127.0.0.1:6379")
	if err != nil {
		fmt.Println("服务器连接异常!")
		this.Redirect("/user/showCart",302)
		return
	}
	defer conn.Close()
	//删除数据
	_,err = conn.Do("hdel","cart_"+userName.(string),id)
	if err != nil {
		fmt.Println("数据库操作异常!")
		this.Redirect("/user/showCart",302)
		return
	}
	//返回数据
	this.Redirect("/user/showCart",302)
}