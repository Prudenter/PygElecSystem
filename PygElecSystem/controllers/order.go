package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"PygElecSystem/PygElecSystem/models"
	"strconv"
	"github.com/gomodule/redigo/redis"
)

type OrderController struct {
	beego.Controller
}

/* 定义函数,负责订单结算页面展示 */
func (this *OrderController)ShowOrder() {
	//获取数据
	goodsIds := this.GetStrings("goodsIds")
	//校验数据
	if len(goodsIds) == 0 {
		this.Redirect("/user/showCart",302)
		return
	}
	//处理数据
	//获取当前用户的所有收货地址
	userName := this.GetSession("userName")
	o := orm.NewOrm()
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",userName.(string)).All(&addrs)
	this.Data["addrs"] = addrs

	//获取当前订单中的所有商品,商品总价,总件数
	//定义总容器
	var goods []map[string]interface{}
	var totalPrice,totalCount int
	var goodsSKU models.GoodsSKU
	//创建redis数据库连接
	conn,_ := redis.Dial("tcp","127.0.0.1:6379")
	for _,v := range goodsIds{
		//定义行容器
		temp := make(map[string]interface{})
		id,_ := strconv.Atoi(v)
		//查询商品信息
		goodsSKU.Id = id
		o.Read(&goodsSKU)

		//获取当前商品在当前订单中的数量
		count,_ := redis.Int(conn.Do("hget","cart_"+userName.(string),id))
		//计算小计
		littlePrice := count * goodsSKU.Price

		//把商品信息放到行容器
		temp["goodsSKU"] = goodsSKU
		temp["littlePrice"] =littlePrice
		temp["count"] = count

		//计算总价和总件数
		totalCount += 1
		totalPrice += littlePrice

		//把行容器加到总容器中
		goods = append(goods,temp)
	}

	//返回数据
	this.Data["totalCount"] = totalCount
	this.Data["totalPrice"] = totalPrice
	//定义邮费
	var postage int = 10
	this.Data["postage"] = postage
	this.Data["actualPrice"] = totalPrice + postage
	this.Data["goods"] = goods
	this.Data["goodsIds"] = goodsIds
	this.TplName = "place_order.html"
}
