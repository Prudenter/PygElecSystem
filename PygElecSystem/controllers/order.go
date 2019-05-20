package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"PygElecSystem/PygElecSystem/models"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"strings"
	"time"
)

type OrderController struct {
	beego.Controller
}

/* 定义函数,负责订单结算页面展示 */
func (this *OrderController) ShowOrder() {
	//获取数据
	goodsIds := this.GetStrings("goodsIds")
	//校验数据
	if len(goodsIds) == 0 {
		this.Redirect("/user/showCart", 302)
		return
	}
	//处理数据
	//获取当前用户的所有收货地址
	userName := this.GetSession("userName")
	o := orm.NewOrm()
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", userName.(string)).All(&addrs)
	this.Data["addrs"] = addrs

	//获取当前订单中的所有商品,商品总价,总件数
	//定义总容器
	var goods []map[string]interface{}
	var totalPrice, totalCount int
	var goodsSKU models.GoodsSKU
	//创建redis数据库连接
	conn, _ := redis.Dial("tcp", "127.0.0.1:6379")
	for _, v := range goodsIds {
		//定义行容器
		temp := make(map[string]interface{})
		id, _ := strconv.Atoi(v)
		//查询商品信息
		goodsSKU.Id = id
		o.Read(&goodsSKU)

		//获取当前商品在当前订单中的数量
		count, _ := redis.Int(conn.Do("hget", "cart_"+userName.(string), id))
		//计算小计
		littlePrice := count * goodsSKU.Price

		//把商品信息放到行容器
		temp["goodsSKU"] = goodsSKU
		temp["littlePrice"] = littlePrice
		temp["count"] = count

		//计算总价和总件数
		totalCount += 1
		totalPrice += littlePrice

		//把行容器加到总容器中
		goods = append(goods, temp)
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

/* 定义函数,负责订单提交业务处理 */
func (this *OrderController) HandleCommitOrder() {
	//获取数据 postage
	addrId, err1 := this.GetInt("addrId")
	payId, err2 := this.GetInt("payId")
	goodsIds := this.GetString("goodsIds")
	totalCount, err3 := this.GetInt("totalCount")
	totalPrice, err4 := this.GetInt("totalPrice")
	postage, err5 := this.GetInt("postage")

	//校验数据,返回json数据给ajax
	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller, resp)

	//校验数据是否完整
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || goodsIds == "" {
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整!"
		return
	}

	//校验用户是否登录
	userName := this.GetSession("userName")
	if userName == nil {
		resp["errno"] = 2
		resp["errmsg"] = "当前用户未登录,请前往登录!"
		return
	}

	//处理数据
	goodsIdSlice := strings.Split(goodsIds[1:len(goodsIds)-1], " ")
	//插入数据,把数据插入到mysql数据库中的订单表和订单商品表
	//获取当前登录的用户对象和地址对象
	var orderInfo models.OrderInfo
	var user models.User
	var address models.Address
	o := orm.NewOrm()
	user.Name = userName.(string)
	o.Read(&user, "Name")
	address.Id = addrId
	o.Read(&address)
	//设置订单号
	orderInfo.OrderId = time.Now().Format("20060102150405" + strconv.Itoa(user.Id))
	orderInfo.User = &user
	orderInfo.Address = &address
	orderInfo.PayMethod = payId
	orderInfo.TotalCount = totalCount
	orderInfo.TotalPrice = totalPrice
	orderInfo.TransitPrice = postage

	//开启mysql事务,防止出现异常
	o.Begin()

	//插入订单表
	o.Insert(&orderInfo)

	//建立redis连接
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		resp["errno"] = 3
		resp["errmsg"] = "网络连接异常!"
		//出现异常,进行事物回滚
		o.Rollback()
		return
	}
	//关闭连接
	defer conn.Close()
	for _, v := range goodsIdSlice {
		//循环遍历所有商品id,查询每个商品在redis和mysql中的商品信息,存入订单商品表
		//每循环一次,重新定义一个orderGoods和goodsSKU,保证每个orderGoods的id不同,如果放在外面又不设置id,那么循环插入时会报错.
		var orderGoods models.OrderGoods
		var goodsSKU models.GoodsSKU

		goodsId, _ := strconv.Atoi(v)
		//获取mysql中的商品信息
		goodsSKU.Id = goodsId
		o.Read(&goodsSKU)
		//获取redis中的商品信息
		count, err := redis.Int(conn.Do("hget", "cart_"+user.Name, goodsId))
		//计算小计
		littlePrice := goodsSKU.Price * count

		if err != nil {
			resp["errno"] = 4
			resp["errmsg"] = "商品不存在!"
			//出现异常,进行事物回滚
			o.Rollback()
			return
		}
		//插入orderGoods
		orderGoods.OrderInfo = &orderInfo
		orderGoods.GoodsSKU = &goodsSKU
		orderGoods.Count = count
		orderGoods.Price = littlePrice

		//校验商品库存和订单中的商品数量
		if goodsSKU.Stock < count {
			resp["errno"] = 6
			resp["errmsg"] = "商品库存不足!"
			//出现异常,进行事物回滚
			o.Rollback()
			return
		}
		//插入之前更新商品库存和销量
		goodsSKU.Stock -= count
		goodsSKU.Sales += count
		o.Update(&goodsSKU)

		//插入orderGoods
		_, err = o.Insert(&orderGoods)
		if err != nil {
			resp["errno"] = 7
			resp["errmsg"] = "订单提交失败!"
			//出现异常,进行事物回滚
			o.Rollback()
			return
		}

		//订单提交成功后,清空购物车中已提交的商品
		_, err = conn.Do("hdel", "cart_"+user.Name, goodsId)
		if err != nil {
			resp["errno"] = 8
			resp["errmsg"] = "清空购物车失败!"
			//出现异常,进行事物回滚
			o.Rollback()
			return
		}
	}
	//提交事务!!
	o.Commit()
	//返回数据
	resp["errno"] = 5
	resp["errmsg"] = "ok!"
}
